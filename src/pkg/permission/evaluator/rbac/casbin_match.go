// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rbac

import (
	"math"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/common/utils/log"
)

type regexpEntry struct {
	data      *regexp.Regexp
	isAliveAt int64
}

func (e *regexpEntry) isAlive() bool {
	return e.isAliveAt >= time.Now().UnixNano()
}

type regexpStore struct {
	aliveDuration time.Duration
	entries       sync.Map
}

func (s *regexpStore) Get(key string, build func(string) *regexp.Regexp) *regexp.Regexp {
	var entry *regexpEntry

	value, ok := s.entries.Load(key)
	if !ok {
		var isAliveAt int64
		if s.aliveDuration > 0 {
			isAliveAt = time.Now().Add(s.aliveDuration).UnixNano()
		} else {
			isAliveAt = math.MaxInt64
		}

		entry = &regexpEntry{
			data:      build(key),
			isAliveAt: isAliveAt,
		}

		s.entries.Store(key, entry)
	} else {
		entry = value.(*regexpEntry)
	}

	return entry.data
}

func (s *regexpStore) Purge() {
	var keys []interface{}
	s.entries.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*regexpEntry); ok && !entry.isAlive() {
			keys = append(keys, key)
		}
		return true
	})

	for _, key := range keys {
		s.entries.Delete(key)
	}
}

var (
	store = &regexpStore{aliveDuration: time.Hour}
	// re = regexp.MustCompile(`(.*):[^/]+(.*)`)
)

func init() {
	startRegexpStorePurging(store, time.Hour)
}

func startRegexpStorePurging(s *regexpStore, intervalDuration time.Duration) {
	go func() {
		rand.Seed(time.Now().Unix())
		jitter := time.Duration(rand.Int()%60) * time.Minute
		log.Debugf("Starting regexp store purge in %s", jitter)
		time.Sleep(jitter)

		for {
			s.Purge()
			log.Debugf("Starting regexp store purge in %s", intervalDuration)
			time.Sleep(intervalDuration)
		}
	}()
}

func keyMatch2Build(key2 string) *regexp.Regexp {
	re := regexp.MustCompile(`(.*):[^/]+(.*)`)

	key2 = strings.Replace(key2, "/*", "/.*", -1)
	for {
		if !strings.Contains(key2, "/:") {
			break
		}

		key2 = re.ReplaceAllString(key2, "$1[^/]+$2")
	}

	return regexp.MustCompile("^" + key2 + "$")
}

// keyMatch2 determines whether key1 matches the pattern of key2, its behavior most likely the builtin KeyMatch2
// except that the match of ("/project/1/robot", "/project/1") will return false
func keyMatch2(key1 string, key2 string) bool {
	return store.Get(key2, keyMatch2Build).MatchString(key1)
}

func keyMatch2Func(args ...interface{}) (interface{}, error) {
	name1 := args[0].(string)
	name2 := args[1].(string)

	return keyMatch2(name1, name2), nil
}
