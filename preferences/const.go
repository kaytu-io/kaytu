package preferences

import (
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type PreferencesYamlFile struct {
	Preferences []PreferenceValueItem
}

type PreferenceValueItem struct {
	Service string
	Key     string
	Value   *string
}

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

func UpdateValues(pis []PreferenceValueItem) error {
	for _, pi := range pis {
		found := false
		for idx, pref := range defaultPref {
			if pref.Service == pi.Service && pref.Key == pi.Key {
				if pi.Value == nil {
					defaultPref[idx].Value = nil
				} else {
					defaultPref[idx].Value = wrapperspb.String(*pi.Value)
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("preferences key %s not found", pi.Key)
		}
	}
	return nil
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
