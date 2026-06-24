package handlers

import (
	"net/http"
	"time"

	"github.com/yimm/rfid-api/config"
	"github.com/yimm/rfid-api/helpers"

	"github.com/gin-gonic/gin"
)

type ScannedItem struct {
	EpcCode string `json:"epc_code"`
	Type    string `json:"type"`
}

type TransactionSubmitRequest struct {
	TransactionDate string        `json:"transaction_date" binding:"required"`
	SupplierCode    string        `json:"supplier_code" binding:"required"`
	LpName          string        `json:"lp_name" binding:"required"`
	RouteCode       string        `json:"route_code" binding:"required"`
	StatusTruck     string        `json:"status_truck" binding:"required"`
	NoTruckEpc      string        `json:"no_truck_epc" binding:"required"`
	ScannedItems    []ScannedItem `json:"scanned_items" binding:"required"`
}

type TransactionSubmitResponse struct {
	LogID  int64 `json:"log_id"`
	Ritase int   `json:"ritase"`
}

type TransactionHistoryItem struct {
	ID              int64         `json:"id"`
	TransactionDate string        `json:"transaction_date"`
	SupplierCode    string        `json:"supplier_code"`
	LpName          string        `json:"lp_name"`
	RouteNo         string        `json:"route_no"`
	StatusTruck     string        `json:"status_truck"`
	NoTruckEpc      string        `json:"no_truck_epc"`
	Ritase          int           `json:"ritase"`
	ScannedItems    []ScannedItem `json:"scanned_items"`
}

func History(c *gin.Context) {
	user, ok := GetCurrentUser(c)
	if !ok {
		return
	}

	type historyRow struct {
		ID              int64
		TransactionDate time.Time
		SupplierCode    string
		LpName          string
		RouteNo         string
		StatusTruck     string
		NoTruckEpc      string
		Ritase          int
	}

	rows, err := config.DB.Query(`
		SELECT id, transaction_date, supplier_code, lp_name, route_no, status_truck, no_truck_epc, ritase
		FROM milk_run_logs
		WHERE user_id = ?
		ORDER BY transaction_date DESC
		LIMIT 15
	`, user.ID)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil riwayat transaksi.")
		return
	}
	defer rows.Close()

	result := []TransactionHistoryItem{}

	for rows.Next() {
		var h historyRow
		if err := rows.Scan(&h.ID, &h.TransactionDate, &h.SupplierCode, &h.LpName, &h.RouteNo, &h.StatusTruck, &h.NoTruckEpc, &h.Ritase); err != nil {
			continue
		}

		itemRows, err := config.DB.Query(`
			SELECT epc_code, item_type
			FROM scanned_items
			WHERE milk_run_log_id = ?
		`, h.ID)
		if err != nil {
			continue
		}

		scannedItems := []ScannedItem{}
		for itemRows.Next() {
			var item ScannedItem
			if err := itemRows.Scan(&item.EpcCode, &item.Type); err != nil {
				continue
			}
			scannedItems = append(scannedItems, item)
		}
		itemRows.Close()

		result = append(result, TransactionHistoryItem{
			ID:              h.ID,
			TransactionDate: h.TransactionDate.Format("02-01-2006 15:04"),
			SupplierCode:    h.SupplierCode,
			LpName:          h.LpName,
			RouteNo:         h.RouteNo,
			StatusTruck:     h.StatusTruck,
			NoTruckEpc:      h.NoTruckEpc,
			Ritase:          h.Ritase,
			ScannedItems:    scannedItems,
		})
	}

	helpers.SuccessDataResponse(c, http.StatusOK, result)
}

func Submit(c *gin.Context) {
	user, ok := GetCurrentUser(c)
	if !ok {
		return
	}

	var req TransactionSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.ValidationErrorResponse(c, http.StatusUnprocessableEntity, "The given data was invalid.", map[string][]string{
			"_": {"The given data was invalid."},
		})
		return
	}

	transactionDate, err := time.Parse("02-01-2006 15:04", req.TransactionDate)
	if err != nil {
		helpers.ValidationErrorResponse(c, http.StatusUnprocessableEntity, "The given data was invalid.", map[string][]string{
			"transaction_date": {"The transaction date does not match the format d-m-Y H:i."},
		})
		return
	}

	var supplierExists bool
	err = config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM master_suppliers WHERE supplier_code = ?)`, req.SupplierCode).Scan(&supplierExists)
	if err != nil || !supplierExists {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Supplier code tidak ditemukan di database.")
		return
	}

	var routeExists bool
	err = config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM master_routes WHERE route_code = ?)`, req.RouteCode).Scan(&routeExists)
	if err != nil || !routeExists {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Route code tidak ditemukan di database.")
		return
	}

	validStatuses := map[string]bool{
		"IN FROM WJ": true, "IN FROM HO": true, "OUT TO WJ": true, "OUT TO HO": true,
	}
	if !validStatuses[req.StatusTruck] {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Status truck tidak valid. Gunakan: IN FROM WJ, IN FROM HO, OUT TO WJ, atau OUT TO HO.")
		return
	}

	var truckExists bool
	err = config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM master_rfid_tags WHERE epc_code = ?)`, req.NoTruckEpc).Scan(&truckExists)
	if err != nil || !truckExists {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "EPC Code Truk tidak terdaftar di sistem.")
		return
	}

	var truckCategory string
	err = config.DB.QueryRow(`SELECT category FROM master_rfid_tags WHERE epc_code = ?`, req.NoTruckEpc).Scan(&truckCategory)
	if err != nil || truckCategory != "Truck" {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "EPC Code yang di-scan bukan kategori Truck.")
		return
	}

	if len(req.ScannedItems) == 0 {
		helpers.ValidationErrorResponse(c, http.StatusUnprocessableEntity, "The given data was invalid.", map[string][]string{
			"scanned_items": {"The scanned items field must have at least 1 item."},
		})
		return
	}

	for _, item := range req.ScannedItems {
		var itemExists bool
		err = config.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM master_rfid_tags WHERE epc_code = ?)`, item.EpcCode).Scan(&itemExists)
		if err != nil || !itemExists {
			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Salah satu EPC Code item (Pallet/Box) tidak valid atau tidak terdaftar.")
			return
		}

		if item.Type != "Pallet" && item.Type != "Box" {
			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "The given data was invalid.")
			return
		}
	}

	dateStr := transactionDate.Format("2006-01-02")
	var ritase int
	err = config.DB.QueryRow(`
		SELECT COUNT(*) + 1 FROM milk_run_logs
		WHERE no_truck_epc = ? AND DATE(transaction_date) = ?
	`, req.NoTruckEpc, dateStr).Scan(&ritase)
	if err != nil {
		ritase = 1
	}

	tx, err := config.DB.Begin()
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Terjadi kesalahan saat menyimpan transaksi.")
		return
	}

	result, err := tx.Exec(`
		INSERT INTO milk_run_logs (user_id, transaction_date, supplier_code, lp_name, route_no, status_truck, no_truck_epc, ritase, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, user.ID, transactionDate, req.SupplierCode, req.LpName, req.RouteCode, req.StatusTruck, req.NoTruckEpc, ritase)
	if err != nil {
		tx.Rollback()
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Terjadi kesalahan saat menyimpan transaksi: "+err.Error())
		return
	}

	logID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Terjadi kesalahan saat menyimpan transaksi.")
		return
	}

	for _, item := range req.ScannedItems {
		_, err = tx.Exec(`
			INSERT INTO scanned_items (milk_run_log_id, epc_code, item_type, created_at, updated_at)
			VALUES (?, ?, ?, NOW(), NOW())
		`, logID, item.EpcCode, item.Type)
		if err != nil {
			tx.Rollback()
			helpers.ErrorResponse(c, http.StatusInternalServerError, "Terjadi kesalahan saat menyimpan transaksi.")
			return
		}
	}

	if err = tx.Commit(); err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Terjadi kesalahan saat menyimpan transaksi.")
		return
	}

	helpers.SuccessResponse(c, http.StatusCreated, "Transaksi berhasil disimpan. Email sedang diproses.", TransactionSubmitResponse{
		LogID:  logID,
		Ritase: ritase,
	})
}