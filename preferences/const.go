package preferences

import (
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
)

var (
	defaultPref []*golang.PreferenceItem
)

func Update(pis []*golang.PreferenceItem) {
	for _, pi := range pis {
		found := false
		for idx, pref := range defaultPref {
			if pref.Service == pi.Service && pref.Key == pi.Key {
				defaultPref[idx] = pi
				found = true
				break
			}
		}
		if !found {
			defaultPref = append(defaultPref, pi)
		}
	}
}

func DefaultPreferences() []*golang.PreferenceItem {
	return defaultPref
}

func Export(pref []*golang.PreferenceItem) map[string]*string {
	ex := map[string]*string{}
	for _, p := range pref {
		if p.Pinned {
			ex[p.Key] = nil
		} else {
			if p.Value != nil {
				ex[p.Key] = &p.Value.Value
			}
		}
	}
	return ex
}
