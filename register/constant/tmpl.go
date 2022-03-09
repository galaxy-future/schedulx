package constant

type ScheduleType string

const (
	ScheduleTypeShrink ScheduleType = "shrink"
	ScheduleTypeExpand ScheduleType = "expand"
	ScheduleTypeDeploy ScheduleType = "deploy"
)
