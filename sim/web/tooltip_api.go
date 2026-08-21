package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/wowsims/wotlk/assets/database"
	"github.com/wowsims/wotlk/sim/core/proto"
)

// tooltipDB is loaded once at startup from the embedded db.bin.
var tooltipDB *proto.UIDatabase

func initTooltipDB() {
	tooltipDB = database.Load()
	log.Printf("Tooltip DB loaded: %d items, %d gems, %d spellIcons",
		len(tooltipDB.Items), len(tooltipDB.Gems), len(tooltipDB.SpellIcons))
}

// tooltipResponse mimics the JSON shape returned by nether.wowhead.com/tooltip.
type tooltipResponse struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

func handleTooltipAPI(w http.ResponseWriter, r *http.Request) {
	// Path format: /wotlk/api/tooltip/{item|spell}/{id}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/wotlk/api/tooltip/"), "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}
	kind := parts[0]
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if tooltipDB == nil {
		initTooltipDB()
	}

	var resp tooltipResponse
	found := false

	switch kind {
	case "item":
		// Search items
		for _, item := range tooltipDB.Items {
			if int(item.Id) == id {
				resp = tooltipResponse{Name: item.Name, Icon: item.Icon}
				found = true
				break
			}
		}
		// If not found in items, search gems (gems have item-like IDs)
		if !found {
			for _, gem := range tooltipDB.Gems {
				if int(gem.Id) == id {
					resp = tooltipResponse{Name: gem.Name, Icon: gem.Icon}
					found = true
					break
				}
			}
		}
	case "spell":
		// Search spellIcons
		for _, spell := range tooltipDB.SpellIcons {
			if int(spell.Id) == id {
				resp = tooltipResponse{Name: spell.Name, Icon: spell.Icon}
				found = true
				break
			}
		}
	}

	if !found {
		resp = tooltipResponse{Name: "", Icon: ""}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// registerTooltipRoutes sets up the local tooltip API routes.
func registerTooltipRoutes() {
	http.HandleFunc("/wotlk/api/tooltip/", func(w http.ResponseWriter, r *http.Request) {
		handleTooltipAPI(w, r)
	})
}
