package handlers

import (
	"database/sql"
	"net/http"

	"github.com/yimm/rfid-api/config"
	"github.com/yimm/rfid-api/helpers"

	"github.com/gin-gonic/gin"
)

type Supplier struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	LpName string `json:"lp_name"`
}

type Route struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type MasterDataResponse struct {
	Suppliers []Supplier `json:"suppliers"`
	Routes    []Route    `json:"routes"`
}

func MasterData(c *gin.Context) {
	suppliers := []Supplier{}
	rows, err := config.DB.Query(`
		SELECT supplier_code, supplier_name, lp_name
		FROM master_suppliers
	`)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data supplier.")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var s Supplier
		if err := rows.Scan(&s.Code, &s.Name, &s.LpName); err != nil {
			continue
		}
		suppliers = append(suppliers, s)
	}

	routes := []Route{}
	rows2, err := config.DB.Query(`
		SELECT route_code, description
		FROM master_routes
	`)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data route.")
		return
	}
	defer rows2.Close()

	for rows2.Next() {
		var r Route
		if err := rows2.Scan(&r.Code, &r.Description); err != nil {
			continue
		}
		routes = append(routes, r)
	}

	helpers.SuccessDataResponse(c, http.StatusOK, MasterDataResponse{
		Suppliers: suppliers,
		Routes:    routes,
	})
}

func MasterRfid(c *gin.Context) {
	type RfidTag struct {
		EpcCode     string `json:"epc_code"`
		Category    string `json:"category"`
		Description string `json:"description"`
	}

	tags := []RfidTag{}
	rows, err := config.DB.Query(`
		SELECT epc_code, category, description
		FROM master_rfid_tags
	`)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data RFID.")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var t RfidTag
		var desc sql.NullString
		if err := rows.Scan(&t.EpcCode, &t.Category, &desc); err != nil {
			continue
		}
		if desc.Valid {
			t.Description = desc.String
		}
		tags = append(tags, t)
	}

	helpers.SuccessDataResponse(c, http.StatusOK, tags)
}
