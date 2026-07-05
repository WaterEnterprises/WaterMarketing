package db

import (
	"bufio"
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

var (
	DBPath     = func() string { p, _ := os.Getwd(); return filepath.Join(p, "databases", "leads.db") }()
	MailDBPath = func() string { p, _ := os.Getwd(); return filepath.Join(p, "databases", "mail-credentials.db") }()
	ENVPath    = func() string {
		p, _ := os.Getwd()
		return filepath.Join(p, ".env")
	}()
)

var Statuses = []string{"cold", "contacted", "replied", "meeting", "negotiating", "closed_won", "closed_lost"}

var Tiers = map[string]string{
	"1": "VC", "2": "Corporate", "3": "Local",
	"4": "Grant", "5": "Venue", "6": "Media",
}

func LoadDBPassword() string {
	if data, err := os.ReadFile(ENVPath); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimRight(line, "\r ")
			if strings.HasPrefix(line, "EMAIL_DB_PASSWORD=") {
				val := line[18:]
				if len(val) >= 2 && (val[0] == '"' || val[0] == '\'') {
					val = val[1 : len(val)-1]
				}
				if val != "" {
					return val
				}
			}
		}
	}
	pw := os.Getenv("EMAIL_DB_PASSWORD")
	if pw != "" {
		return pw
	}
	fmt.Println("ERROR: EMAIL_DB_PASSWORD not found in .env or environment")
	os.Exit(1)
	return ""
}

func openDB(path string) (*sql.DB, error) {
	pw := LoadDBPassword()
	hexKey := fmt.Sprintf("%x", []byte(pw))
	dsn := fmt.Sprintf("%s?_pragma_key=x'%s'", path, hexKey)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA foreign_keys=ON")
	return db, nil
}

func GetDB() (*sql.DB, error)  { return openDB(DBPath) }
func GetMailDB() (*sql.DB, error) { return openDB(MailDBPath) }

func InitDB() error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS leads (
			id TEXT PRIMARY KEY,
			company TEXT NOT NULL UNIQUE,
			contact_name TEXT DEFAULT '',
			email TEXT DEFAULT '',
			phone TEXT DEFAULT '',
			website TEXT DEFAULT '',
			tier TEXT NOT NULL DEFAULT '3',
			type TEXT NOT NULL DEFAULT '',
			vertical TEXT DEFAULT '',
			check_size TEXT DEFAULT '',
			pitch_angle TEXT DEFAULT '',
			status TEXT DEFAULT 'cold',
			next_action TEXT DEFAULT '',
			next_action_date TEXT DEFAULT '',
			notes TEXT DEFAULT '',
			source TEXT DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		);
		CREATE TABLE IF NOT EXISTS outreach_log (
			id TEXT PRIMARY KEY,
			lead_id TEXT NOT NULL REFERENCES leads(id) ON DELETE CASCADE,
			activity_type TEXT NOT NULL,
			notes TEXT DEFAULT '',
			outcome TEXT DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		);`)
	if err != nil {
		return err
	}
	// add UNIQUE index on company if table already existed without it
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_leads_company ON leads(company)")
	return nil
}

type Lead struct {
	ID             string `json:"id"`
	Company        string `json:"company"`
	ContactName    string `json:"contact_name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Website        string `json:"website"`
	Tier           string `json:"tier"`
	Type           string `json:"type"`
	Vertical       string `json:"vertical"`
	CheckSize      string `json:"check_size"`
	PitchAngle     string `json:"pitch_angle"`
	Status         string `json:"status"`
	NextAction     string `json:"next_action"`
	NextActionDate string `json:"next_action_date"`
	Notes          string `json:"notes"`
	Source         string `json:"source"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type OutreachEntry struct {
	ID           string `json:"id"`
	LeadID       string `json:"lead_id"`
	ActivityType string `json:"activity_type"`
	Notes        string `json:"notes"`
	Outcome      string `json:"outcome"`
	CreatedAt    string `json:"created_at"`
}

type Stats struct {
	Total        int           `json:"total"`
	ByTier       []TierCount   `json:"by_tier"`
	ByStatus     []StatusCount `json:"by_status"`
	FollowupsDue int           `json:"followups_due"`
	Recent       []Lead        `json:"recent"`
}

type TierCount struct {
	Tier  string `json:"tier"`
	Count int    `json:"count"`
}

type StatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type PaginatedLeads struct {
	Data  []Lead `json:"data"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}

func uuid() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func scanLead(scanner interface{ Scan(...interface{}) error }) (*Lead, error) {
	var l Lead
	err := scanner.Scan(&l.ID, &l.Company, &l.ContactName, &l.Email, &l.Phone,
		&l.Website, &l.Tier, &l.Type, &l.Vertical, &l.CheckSize,
		&l.PitchAngle, &l.Status, &l.NextAction, &l.NextActionDate,
		&l.Notes, &l.Source, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func scanLeads(rows *sql.Rows) ([]Lead, error) {
	var leads []Lead
	for rows.Next() {
		l, err := scanLead(rows)
		if err != nil {
			return nil, err
		}
		leads = append(leads, *l)
	}
	return leads, rows.Err()
}

func scanOutreach(rows *sql.Rows) ([]OutreachEntry, error) {
	var entries []OutreachEntry
	for rows.Next() {
		var e OutreachEntry
		if err := rows.Scan(&e.ID, &e.LeadID, &e.ActivityType, &e.Notes, &e.Outcome, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func AddLead(data Lead) (string, error) {
	if data.Company == "" {
		return "", fmt.Errorf("company is required")
	}
	if data.Tier == "" {
		return "", fmt.Errorf("tier is required")
	}
	if data.Type == "" {
		return "", fmt.Errorf("type is required")
	}
	id := uuid()
	db, err := GetDB()
	if err != nil {
		return "", err
	}
	defer db.Close()

	_, err = db.Exec(`INSERT INTO leads (id, company, contact_name, email, phone, website,
		tier, type, vertical, check_size, pitch_angle, status,
		next_action, next_action_date, notes, source)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, data.Company, data.ContactName, data.Email, data.Phone,
		data.Website, data.Tier, data.Type, data.Vertical, data.CheckSize,
		data.PitchAngle, data.Status, data.NextAction, data.NextActionDate,
		data.Notes, data.Source)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return "", fmt.Errorf("a lead with company '%s' already exists", data.Company)
		}
		return "", err
	}
	return id, nil
}

func GetLeads(filters map[string]string) (PaginatedLeads, error) {
	db, err := GetDB()
	if err != nil {
		return PaginatedLeads{}, err
	}
	defer db.Close()

	clauses := []string{}
	params := []interface{}{}

	if t, ok := filters["tier"]; ok && t != "" {
		clauses = append(clauses, "tier = ?")
		params = append(params, t)
	}
	if s, ok := filters["status"]; ok && s != "" {
		if s == "active" {
			clauses = append(clauses, "status NOT IN ('closed_won','closed_lost')")
		} else {
			clauses = append(clauses, "status = ?")
			params = append(params, s)
		}
	}
	if v, ok := filters["vertical"]; ok && v != "" {
		clauses = append(clauses, "vertical = ?")
		params = append(params, v)
	}
	if t, ok := filters["type"]; ok && t != "" {
		clauses = append(clauses, "type = ?")
		params = append(params, t)
	}
	if search, ok := filters["search"]; ok && search != "" {
		s := "%" + search + "%"
		clauses = append(clauses, "(company LIKE ? OR contact_name LIKE ? OR email LIKE ?)")
		params = append(params, s, s, s)
	}

	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}

	page := 1
	limit := 50
	if p, ok := filters["page"]; ok && p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if l, ok := filters["limit"]; ok && l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	offset := (page - 1) * limit

	var total int
	countQuery := "SELECT COUNT(*) FROM leads " + where
	db.QueryRow(countQuery, params...).Scan(&total)

	rows, err := db.Query("SELECT * FROM leads "+where+" ORDER BY updated_at DESC LIMIT ? OFFSET ?",
		append(params, limit, offset)...)
	if err != nil {
		return PaginatedLeads{}, err
	}
	defer rows.Close()
	data, err := scanLeads(rows)
	if err != nil {
		return PaginatedLeads{}, err
	}
	if data == nil {
		data = []Lead{}
	}
	return PaginatedLeads{Data: data, Total: total, Page: page, Limit: limit}, nil
}

func GetLead(id string) (*Lead, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM leads WHERE id = ?", id)
	l, err := scanLead(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return l, nil
}

func UpdateLead(id string, data map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	sets := []string{}
	params := []interface{}{}
	for k, v := range data {
		if k == "id" || k == "created_at" {
			continue
		}
		sets = append(sets, k+" = ?")
		params = append(params, v)
	}
	if len(sets) == 0 {
		return nil
	}
	sets = append(sets, "updated_at = datetime('now')")
	params = append(params, id)

	_, err = db.Exec("UPDATE leads SET "+strings.Join(sets, ", ")+" WHERE id = ?", params...)
	return err
}

func DeleteLead(id string) error {
	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	db.Exec("DELETE FROM outreach_log WHERE lead_id = ?", id)
	_, err = db.Exec("DELETE FROM leads WHERE id = ?", id)
	return err
}

func LogOutreach(leadID, activityType, notes, outcome string) (string, error) {
	id := uuid()
	db, err := GetDB()
	if err != nil {
		return "", err
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO outreach_log (id, lead_id, activity_type, notes, outcome) VALUES (?, ?, ?, ?, ?)",
		id, leadID, activityType, notes, outcome)
	if err != nil {
		return "", err
	}
	return id, nil
}

func GetOutreach(leadID string) ([]OutreachEntry, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM outreach_log WHERE lead_id = ? ORDER BY created_at DESC", leadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOutreach(rows)
}

func GetStats() (*Stats, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	s := &Stats{}

	db.QueryRow("SELECT COUNT(*) FROM leads").Scan(&s.Total)

	rows, err := db.Query("SELECT tier, COUNT(*) FROM leads GROUP BY tier")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tc TierCount
			rows.Scan(&tc.Tier, &tc.Count)
			s.ByTier = append(s.ByTier, tc)
		}
	}

	rows2, err := db.Query("SELECT status, COUNT(*) FROM leads GROUP BY status")
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var sc StatusCount
			rows2.Scan(&sc.Status, &sc.Count)
			s.ByStatus = append(s.ByStatus, sc)
		}
	}

	db.QueryRow(`SELECT COUNT(*) FROM leads WHERE next_action_date != ''
		AND next_action_date <= date('now')
		AND status NOT IN ('closed_won','closed_lost')`).Scan(&s.FollowupsDue)

	rows3, err := db.Query("SELECT * FROM leads ORDER BY created_at DESC LIMIT 5")
	if err == nil {
		defer rows3.Close()
		s.Recent, _ = scanLeads(rows3)
	}

	return s, nil
}

func parseCSVLine(line string) []string {
	var fields []string
	inQuotes := false
	current := ""
	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == '"' {
			if inQuotes && i+1 < len(line) && line[i+1] == '"' {
				current += "\""
				i++
			} else {
				inQuotes = !inQuotes
			}
		} else if c == ',' && !inQuotes {
			fields = append(fields, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	fields = append(fields, current)
	return fields
}

func ImportCSV(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("file not found: %s", path)
	}
	defer f.Close()

	db, err := GetDB()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	scanner := bufio.NewScanner(f)

	if !scanner.Scan() {
		return 0, fmt.Errorf("empty CSV")
	}
	headers := parseCSVLine(scanner.Text())

	headerIdx := func(name string) int {
		lower := strings.ToLower(strings.TrimSpace(name))
		for i, h := range headers {
			if strings.EqualFold(strings.TrimSpace(h), lower) {
				return i
			}
		}
		for i, h := range headers {
			if strings.Contains(strings.ToLower(h), lower) {
				return i
			}
		}
		return -1
	}

	count := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := parseCSVLine(line)
		get := func(name string) string {
			idx := headerIdx(name)
			if idx >= 0 && idx < len(fields) {
				return strings.TrimSpace(fields[idx])
			}
			return ""
		}

		company := get("Company")
		if company == "" {
			continue
		}

		tier := get("Tier")
		if tier == "" {
			tier = "3"
		} else {
			tier = string(tier[0])
		}

		emailSent := get("Email Sent")
		if emailSent == "" {
			emailSent = get("Email Sent (Date)")
		}

		pitchAngle := get("Our Angle")
		if pitchAngle == "" {
			pitchAngle = get("Pitch Angle")
		}

		contactName := get("Contact Name")
		if contactName == "" {
			contactName = get("Contact")
		}

		email := get("Email")
		if email == "" {
			email = get("Contact Email")
		}

		id := uuid()
		_, err := db.Exec(`INSERT INTO leads (id, company, contact_name, email, phone, website,
			tier, type, vertical, check_size, pitch_angle, status,
			next_action, next_action_date, notes, source)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, company, contactName, email, get("Phone"), get("Website"),
			tier, get("Type"), get("Vertical"), get("Check Size"), pitchAngle,
			"cold", get("Next Action"), emailSent, get("Notes"), "csv_import")
		if err != nil {
			continue
		}
		count++
	}

	return count, scanner.Err()
}
