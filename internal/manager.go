package internal

import (
	"sync"
)

type ServiceItem struct {
	Name           string `json:"name"`
	SwaggerVersion string `json:"swaggerVersion"`
	Url            string `json:"url"`
	Location       string `json:"location"`
	Host           string `json:"host"`
	Key            string `json:"key"`
}

type ModuleServiceItems map[string]*ServiceItem // host:item

type Manager struct {
	services sync.Map //map[string]ModuleServiceItems // model :{host:item,...}
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) DocServices(proxyRoutePrefix string) []*ServiceItem {
	var data []*ServiceItem

	m.services.Range(func(key, value interface{}) bool {
		for _, item := range value.(ModuleServiceItems) {
			data = append(data, &ServiceItem{
				Name:           item.Name,
				SwaggerVersion: item.SwaggerVersion,
				Url:            proxyRoutePrefix + "/" + key.(string),
				Location:       proxyRoutePrefix + "/" + key.(string),
				Host:           item.Host,
				Key:            key.(string),
			})
			break
		}
		return true
	})
	return data
}

func (m *Manager) Add(module, host, url, debugOrigin, name string) {
	var mdItems ModuleServiceItems
	if v, ok := m.services.Load(module); !ok {
		mdItems = make(ModuleServiceItems)
	} else {
		mdItems = v.(ModuleServiceItems)
	}

	if _, ok := mdItems[host]; ok {
		if url != "" {
			mdItems[host].Url = url
			mdItems[host].Location = url
		}
		if debugOrigin != "" {
			mdItems[host].Host = debugOrigin
		}
		if name != "" {
			mdItems[host].Name = name
		}
	} else {
		mdItems[host] = &ServiceItem{
			Name:           name,
			SwaggerVersion: "2.0",
			Url:            url,
			Location:       url,
			Host:           debugOrigin,
		}
	}
	m.services.Store(module, mdItems)
}

func (m *Manager) Remove(module, host string) {
	if v, ok := m.services.Load(module); ok {
		mdItems := v.(ModuleServiceItems)
		if _, ok = mdItems[host]; ok {
			delete(mdItems, host)
		}
		if len(mdItems) == 0 {
			m.services.Delete(module)
		}
	}
}

func (m *Manager) GetModuleDocUrl(module string) string {
	if v, ok := m.services.Load(module); ok {
		for _, item := range v.(ModuleServiceItems) {
			return item.Location
		}
	}

	return ""
}
