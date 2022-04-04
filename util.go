package panylcli

// PluginsEnabledUnique removes duplicate items from the list
func PluginsEnabledUnique(pluginsEnabled []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range pluginsEnabled {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
