package database

import (
	"database/sql"
	"fmt"
	"regexp"
	"slices"
	"sync"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/umahmood/haversine"
)

var regexCache = struct {
	sync.RWMutex
	toDelete chan string
	cache    map[string]*regexp.Regexp
}{
	cache:    make(map[string]*regexp.Regexp),
	toDelete: make(chan string),
}

func cachedRegex(pattern string) (*regexp.Regexp, error) {
	regexCache.RLock()
	re, exists := regexCache.cache[pattern]
	regexCache.RUnlock()

	if !exists {
		var err error
		re, err = regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		regexCache.Lock()
		regexCache.cache[pattern] = re
		regexCache.Unlock()
		go func() {
			// delete cached compiled regex after timeout
			time.Sleep(10 * time.Second)
			regexCache.toDelete <- pattern
		}()
	}

	return re, nil
}

func regexMatch(pattern, text string) (bool, error) {
	re, err := cachedRegex(pattern)
	if err != nil {
		return false, err
	}
	if re.MatchString(text) {
		return true, nil
	}
	return false, nil
}

func regexCapture(pattern, text string) (string, error) {
	re, err := cachedRegex(pattern)
	if err != nil {
		return "", err
	}
	match := re.FindStringSubmatch(text)
	if len(match) != 2 {
		return "", fmt.Errorf("no capture groups")
	}
	return match[1], nil
}

func cosineDistanceDegrees(latA float64, longA float64, latB float64, longB float64) float64 {
	mi, _ := haversine.Distance(haversine.Coord{Lat: latA, Lon: longA}, haversine.Coord{Lat: latB, Lon: longB})
	return mi / 69.1
}

func isSqliteRegistered(name string) bool {
	drivers := sql.Drivers()
	return slices.Contains(drivers, name)
}

func registerExtendedSqlite(name string) {
	if isSqliteRegistered(name) {
		return
	}
	go func() {
		// delete cached compiled regex after timeout
		toDelete := <-regexCache.toDelete
		regexCache.Lock()
		delete(regexCache.cache, toDelete)
		regexCache.Unlock()
	}()
	sql.Register(name, &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) (err error) {
			if err = conn.RegisterFunc("regexp", regexMatch, true); err != nil {
				return
			}
			if err = conn.RegisterFunc("gcirc", cosineDistanceDegrees, true); err != nil {
				return
			}
			if err = conn.RegisterFunc("rextract", regexCapture, true); err != nil {
				return
			}
			return
		},
	})
}
