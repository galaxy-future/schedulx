package zadigcli

const (
	listWorkflowUrl      = "/api/aslan/workflow/workflow"
	listWorkflowTasksUrl = "/api/aslan/workflow/workflowtask/max/%d/start/%d/pipelines/%s"
	listS3storageUrl     = "/api/aslan/system/s3storage"
)

const (
	taskBuild          = "buildv2"
	taskDistributeToS3 = "distribute2kodo"

	fileTypeFile  = "file"
	fileTypeImage = "image"
)
