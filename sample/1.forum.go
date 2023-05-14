package sample

import (
	"embed"
	"log"

	"esbconcept/components"
	"esbconcept/esbapi"
	"esbconcept/esbcore"
)

//go:embed flowmodel.*.toml
var flowModelFs embed.FS

//go:embed flow.*.toml
var flowFs embed.FS

//go:embed pipeline.*.toml
var pipelineFs embed.FS

func StartForum() error {
	container := esbcore.NewContainer()
	if err := components.InitComponent(container); err != nil {
		return err
	}
	if err := loadCustomFn(container, map[string]esbapi.FnGen{
		"#print_obj": FnPrintObject,
	}); err != nil {
		return err
	}

	if err := loadFlowModel(container, []string{
		"flowmodel.user.toml",
	}); err != nil {
		return err
	}
	if err := loadFlow(container, map[string]string{
		"register_validation":        "flow.register.validation.toml",
		"send_register_notification": "flow.register.send_notification.toml",
	}); err != nil {
		return err
	}
	if err := loadPipeline(container, map[string]string{
		"register": "pipeline.register.toml",
	}); err != nil {
		return err
	}

	if err := container.StartContainer(); err != nil {
		return err
	}

	return nil
}

func loadCustomFn(container esbapi.Container, mapping map[string]esbapi.FnGen) error {
	for name, fg := range mapping {
		if err := container.RegisterCustomFn(name, fg); err != nil {
			return err
		}
	}
	return nil
}

func loadFlowModel(container *esbcore.ContainerInst, files []string) error {
	for _, file := range files {
		data, err := flowModelFs.ReadFile(file)
		if err != nil {
			return err
		}
		log.Println("read FlowModel content:", string(data))

		if err := container.LoadFlowModel(string(data)); err != nil {
			return err
		}
	}
	return nil
}

func loadFlow(container *esbcore.ContainerInst, flowFiles map[string]string) error {
	for flowName, file := range flowFiles {
		data, err := flowFs.ReadFile(file)
		if err != nil {
			return err
		}
		log.Println("read Flow content:", string(data))

		if err := container.LoadFlow(flowName, string(data)); err != nil {
			return err
		}
	}
	return nil
}

func loadPipeline(container *esbcore.ContainerInst, pipelineFiles map[string]string) error {
	for pipelineName, file := range pipelineFiles {
		data, err := pipelineFs.ReadFile(file)
		if err != nil {
			return err
		}
		log.Println("read Pipeline content:", string(data))

		if err := container.LoadPipeline(pipelineName, string(data)); err != nil {
			return err
		}
	}
	return nil
}
