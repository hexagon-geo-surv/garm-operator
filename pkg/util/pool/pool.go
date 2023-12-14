package pool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cloudbase/garm/client/enterprises"
	"github.com/cloudbase/garm/client/organizations"
	"github.com/cloudbase/garm/client/pools"
	"github.com/cloudbase/garm/client/repositories"
	"github.com/cloudbase/garm/params"
	"sigs.k8s.io/controller-runtime/pkg/log"

	garmoperatorv1alpha1 "github.com/mercedes-benz/garm-operator/api/v1alpha1"
	garmClient "github.com/mercedes-benz/garm-operator/pkg/client"
	"github.com/mercedes-benz/garm-operator/pkg/filter"
	poolfilter "github.com/mercedes-benz/garm-operator/pkg/filter/pool"
)

func GetGarmPoolBySpecs(ctx context.Context, garmClient garmClient.PoolClient, pool *garmoperatorv1alpha1.Pool, image *garmoperatorv1alpha1.Image, gitHubScopeRef garmoperatorv1alpha1.GitHubScope) (params.Pool, error) {
	log := log.FromContext(ctx)
	log.Info("Getting existing garm pools by pool.spec")

	id := gitHubScopeRef.GetID()
	scope, err := garmoperatorv1alpha1.ToGitHubScopeKind(gitHubScopeRef.GetKind())
	if err != nil {
		return params.Pool{}, err
	}

	garmPools, err := garmClient.ListAllPools(pools.NewListPoolsParams())
	if err != nil {
		return params.Pool{}, err
	}
	filteredGarmPools := filter.Match(garmPools.Payload,
		poolfilter.MatchesImage(image.Spec.Tag),
		poolfilter.MatchesFlavor(pool.Spec.Flavor),
		poolfilter.MatchesProvider(pool.Spec.ProviderName),
		poolfilter.MatchesGitHubScope(scope, id),
	)

	if len(filteredGarmPools) > 1 {
		return params.Pool{}, errors.New("can not create pool, multiple instances matching flavour, image and provider found in garm")
	}

	// sync
	if len(filteredGarmPools) == 1 {
		return filteredGarmPools[0], nil
	}

	// create
	return params.Pool{}, nil
}

func UpdatePool(ctx context.Context, garmClient garmClient.PoolClient, pool *garmoperatorv1alpha1.Pool, image *garmoperatorv1alpha1.Image) error {
	log := log.FromContext(ctx).
		WithName("UpdatePool")

	log.Info("updating pool", "pool", pool.Name, "id", pool.Status.ID)

	poolParams := params.UpdatePoolParams{
		RunnerPrefix: params.RunnerPrefix{
			Prefix: pool.Spec.RunnerPrefix,
		},
		MaxRunners:             &pool.Spec.MaxRunners,
		MinIdleRunners:         &pool.Spec.MinIdleRunners,
		Image:                  image.Spec.Tag,
		Flavor:                 pool.Spec.Flavor,
		OSType:                 pool.Spec.OSType,
		OSArch:                 pool.Spec.OSArch,
		Tags:                   pool.Spec.Tags,
		Enabled:                &pool.Spec.Enabled,
		RunnerBootstrapTimeout: &pool.Spec.RunnerBootstrapTimeout,
		ExtraSpecs:             json.RawMessage([]byte(pool.Spec.ExtraSpecs)),
		GitHubRunnerGroup:      &pool.Spec.GitHubRunnerGroup,
	}
	_, err := garmClient.UpdatePool(pools.NewUpdatePoolParams().WithPoolID(pool.Status.ID).WithBody(poolParams))
	if err != nil {
		return err
	}

	return nil
}

func CreatePool(ctx context.Context, garmClient garmClient.PoolClient, pool *garmoperatorv1alpha1.Pool, image *garmoperatorv1alpha1.Image, gitHubScopeRef garmoperatorv1alpha1.GitHubScope) (params.Pool, error) {
	log := log.FromContext(ctx).
		WithName("CreatePool")
	log.Info("creating pool", "pool", pool.Name)

	poolResult := params.Pool{}

	id := gitHubScopeRef.GetID()
	scope, err := garmoperatorv1alpha1.ToGitHubScopeKind(gitHubScopeRef.GetKind())
	if err != nil {
		return poolResult, err
	}

	extraSpecs := json.RawMessage([]byte{})
	if pool.Spec.ExtraSpecs != "" {
		err := json.Unmarshal([]byte(pool.Spec.ExtraSpecs), &extraSpecs)
		if err != nil {
			return poolResult, err
		}
	}

	poolParams := params.CreatePoolParams{
		RunnerPrefix: params.RunnerPrefix{
			Prefix: pool.Spec.RunnerPrefix,
		},
		ProviderName:           pool.Spec.ProviderName,
		MaxRunners:             pool.Spec.MaxRunners,
		MinIdleRunners:         pool.Spec.MinIdleRunners,
		Image:                  image.Spec.Tag,
		Flavor:                 pool.Spec.Flavor,
		OSType:                 pool.Spec.OSType,
		OSArch:                 pool.Spec.OSArch,
		Tags:                   pool.Spec.Tags,
		Enabled:                pool.Spec.Enabled,
		RunnerBootstrapTimeout: pool.Spec.RunnerBootstrapTimeout,
		ExtraSpecs:             extraSpecs,
		GitHubRunnerGroup:      pool.Spec.GitHubRunnerGroup,
	}

	switch scope {
	case garmoperatorv1alpha1.EnterpriseScope:
		result, err := garmClient.CreateEnterprisePool(
			enterprises.NewCreateEnterprisePoolParams().
				WithEnterpriseID(id).
				WithBody(poolParams))
		if err != nil {
			return params.Pool{}, err
		}

		poolResult = result.Payload
	case garmoperatorv1alpha1.OrganizationScope:
		result, err := garmClient.CreateOrgPool(
			organizations.NewCreateOrgPoolParams().
				WithOrgID(id).
				WithBody(poolParams))
		if err != nil {
			return params.Pool{}, err
		}
		poolResult = result.Payload
	case garmoperatorv1alpha1.RepositoryScope:
		result, err := garmClient.CreateRepoPool(
			repositories.NewCreateRepoPoolParams().
				WithRepoID(id).
				WithBody(poolParams))
		if err != nil {
			return params.Pool{}, err
		}
		poolResult = result.Payload
	default:
		err := fmt.Errorf("no valid scope specified: %s", scope)
		return params.Pool{}, err
	}

	return poolResult, nil
}