package cfnpatcher

import (
	"github.com/Jeffail/gabs/v2"
	"github.com/falcosecurity/kilt/pkg/kilt"
	"github.com/rs/zerolog/log"
	"os"
)

func extractContainerInfo(group *gabs.Container, groupName string, container *gabs.Container, configuration *Configuration) *kilt.TargetInfo {
	info := new(kilt.TargetInfo)

	info.ContainerName = container.S("Name").Data().(string)
	info.ContainerGroupName = groupName
	info.EnvironmentVariables = make(map[string]string)
	info.Metadata = make(map[string]string)

	if container.Exists("Image") {
		info.Image = container.S("Image").Data().(string)
		os.Setenv("HOME", "/tmp")  // crane requires $HOME variable
		repoInfo, err := getConfigFromRepository(info.Image)
		if err != nil {
			log.Warn().Str("image", info.Image).Err(err).Msg("could not retrieve metadata from repository")
		}else{
			if configuration.UseRepositoryHints {
				info.EntryPoint = repoInfo.Entrypoint
				info.Command = repoInfo.Command
			}
		}
	}

	if container.Exists("EntryPoint") {
		info.EntryPoint = make([]string, 0)
		for _, arg := range container.S("EntryPoint").Children() {
			info.EntryPoint = append(info.EntryPoint, arg.Data().(string))
		}
	}else{
		log.Warn().Str("image", info.Image).Msg("no EntryPoint was specified")
	}

	if container.Exists("Command") {
		info.Command = make([]string, 0)
		for _, arg := range container.S("Command").Children() {
			info.Command = append(info.Command, arg.Data().(string))
		}
	}else{
		log.Warn().Str("image", info.Image).Msg("no Command was specified")
	}

	if container.Exists("Environment") {
		for _, env := range container.S("Environment").Children() {
			info.EnvironmentVariables[env.S("Name").Data().(string)] = env.S("Value").Data().(string)
		}
	}

	if group.Exists("Properties", "Tags") {
		for _, tag := range group.S("Properties", "Tags").Children() {
			if tag.Exists("Key") && tag.Exists("Value") {
				info.Metadata[tag.S("Key").Data().(string)] = tag.S("Value").Data().(string)
			}
		}
	}

	// TODO(admiral0): metadata tags

	return info
}
