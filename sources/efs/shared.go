package efs

import (
	"github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/overmindtech/sdp-go"
)

// lifeCycleStateToHealth Converts a lifecycle state to a health state
func lifeCycleStateToHealth(state types.LifeCycleState) *sdp.Health {
	switch state {
	case types.LifeCycleStateCreating:
		return sdp.Health_HEALTH_PENDING.Enum()
	case types.LifeCycleStateAvailable:
		return sdp.Health_HEALTH_OK.Enum()
	case types.LifeCycleStateUpdating:
		return sdp.Health_HEALTH_PENDING.Enum()
	case types.LifeCycleStateDeleting:
		return sdp.Health_HEALTH_PENDING.Enum()
	case types.LifeCycleStateDeleted:
		return sdp.Health_HEALTH_WARNING.Enum()
	case types.LifeCycleStateError:
		return sdp.Health_HEALTH_ERROR.Enum()
	}

	return nil
}
