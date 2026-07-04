package main

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"waterenterprises/internal/db"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"golang.org/x/term"
)

func main() {
	if err := db.InitDB(); err != nil {
		fmt.Fprintf(os.Stderr, "DB init error: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		cmdStats()
		return
	}

	command := os.Args[1]
	switch command {
	case "serve":
		cmdServe()
	case "stats":
		cmdStats()
	case "list":
		cmdList()
	case "view":
		cmdView()
	case "add":
		cmdAdd()
	case "update":
		cmdUpdate()
	case "delete":
		cmdDelete()
	case "status":
		cmdStatus()
	case "log":
		cmdLog()
	case "followups":
		cmdFollowups()
	case "import":
		cmdImport()
	case "export":
		cmdExport()
	case "mail":
		cmdMail()
	case "password":
		cmdPassword()
	case "telegram":
		cmdTelegram()
	case "campaign":
		cmdCampaign()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Commands:")
		fmt.Println("  CRM:       stats, list, view, add, update, delete, status,")
		fmt.Println("             log, followups, import, export")
		fmt.Println("  Server:    serve")
		fmt.Println("  Email:     mail, password")
		fmt.Println("  Other:     telegram, campaign")
		os.Exit(1)
	}
}

// ─── Helpers ───────────────────────────────────────────────────────────────

func getFlag(name string) string {
	for i, arg := range os.Args {
		if (arg == "--"+name || arg == "-"+name) && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
		if strings.HasPrefix(arg, "--"+name+"=") {
			return strings.SplitN(arg, "=", 2)[1]
		}
	}
	return ""
}

func hasFlag(name string) bool {
	for _, arg := range os.Args {
		if arg == "--"+name || arg == "-"+name {
			return true
		}
	}
	return false
}

func prompt(label string) string {
	fmt.Print("  " + label + ": ")
	var val string
	fmt.Scanln(&val)
	return strings.TrimSpace(val)
}

// ─── CRM Display ───────────────────────────────────────────────────────────

func printLead(l db.Lead, verbose bool) {
	tierLabel := db.Tiers[l.Tier]
	if tierLabel == "" {
		tierLabel = "Tier " + l.Tier
	}
	shortID := l.ID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	fmt.Printf("  [%s] %s\n", shortID, l.Company)
	contact := ifEmpty(l.ContactName, "-")
	email := ifEmpty(l.Email, "-")
	fmt.Printf("         Contact: %s  |  %s\n", contact, email)
	ltype := ifEmpty(l.Type, "-")
	vert := ifEmpty(l.Vertical, "-")
	fmt.Printf("         Tier: %s  |  Type: %s  |  Vertical: %s\n", tierLabel, ltype, vert)
	cs := ifEmpty(l.CheckSize, "-")
	fmt.Printf("         Status: %s  |  Check: %s\n", l.Status, cs)
	if l.NextAction != "" {
		nd := ifEmpty(l.NextActionDate, "no date")
		fmt.Printf("         Next: %s  (%s)\n", l.NextAction, nd)
	}
	if verbose {
		fmt.Printf("         Website: %s\n", ifEmpty(l.Website, "-"))
		fmt.Printf("         Phone: %s\n", ifEmpty(l.Phone, "-"))
		fmt.Printf("         Pitch: %s\n", ifEmpty(l.PitchAngle, "-"))
		fmt.Printf("         Notes: %s\n", ifEmpty(l.Notes, "-"))
		fmt.Printf("         Created: %s  |  Updated: %s\n", l.CreatedAt, l.UpdatedAt)
		outreach, err := db.GetOutreach(l.ID)
		if err == nil && len(outreach) > 0 {
			fmt.Printf("         Activity (%d):\n", len(outreach))
			max := 5
			if len(outreach) < max {
				max = len(outreach)
			}
			for _, o := range outreach[:max] {
				datePart := o.CreatedAt
				if len(datePart) > 10 {
					datePart = datePart[:10]
				}
				fmt.Printf("           [%s] %s: %s\n", datePart, o.ActivityType, o.Notes)
			}
		}
	}
	fmt.Println()
}

func ifEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// ─── CRM Commands ──────────────────────────────────────────────────────────

func cmdStats() {
	s, err := db.GetStats()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	fmt.Printf("\n%s\n", strings.Repeat("=", 50))
	fmt.Println("  WATERPARTY CRM — DASHBOARD")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("  Total leads:     %d\n", s.Total)
	fmt.Printf("  Follow-ups due:  %d\n", s.FollowupsDue)
	fmt.Println()
	fmt.Println("  By Tier:")
	for _, t := range s.ByTier {
		label := ifEmpty(db.Tiers[t.Tier], "Tier "+t.Tier)
		fmt.Printf("    %s: %d\n", label, t.Count)
	}
	fmt.Println()
	fmt.Println("  By Status:")
	for _, st := range s.ByStatus {
		fmt.Printf("    %s: %d\n", st.Status, st.Count)
	}
	fmt.Println()
	if len(s.Recent) > 0 {
		fmt.Println("  Recent:")
		for _, r := range s.Recent {
			shortID := r.ID
			if len(shortID) > 8 {
				shortID = shortID[:8]
			}
			fmt.Printf("    [%s] %s — %s\n", shortID, r.Company, r.Status)
		}
	}
	fmt.Println()
}

func cmdList() {
	filters := map[string]string{}
	if t := getFlag("tier"); t != "" {
		filters["tier"] = t
	}
	if s := getFlag("status"); s != "" {
		filters["status"] = s
	}
	if s := getFlag("search"); s != "" {
		filters["search"] = s
	}
	if v := getFlag("vertical"); v != "" {
		filters["vertical"] = v
	}
	if t := getFlag("type"); t != "" {
		filters["type"] = t
	}
	leads, err := db.GetLeads(filters)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	if len(leads) == 0 {
		fmt.Println("No leads found.")
		return
	}
	fmt.Printf("\n%d lead(s):\n\n", len(leads))
	verbose := hasFlag("verbose") || hasFlag("v")
	for _, l := range leads {
		printLead(l, verbose)
	}
}

func cmdView() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run crm.go view <id>")
		return
	}
	id := os.Args[2]
	l, err := db.GetLead(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	if l == nil {
		fmt.Println("Lead not found.")
		return
	}
	printLead(*l, true)
	outreach, err := db.GetOutreach(id)
	if err != nil {
		return
	}
	if len(outreach) > 0 {
		fmt.Printf("  All Activity (%d):\n", len(outreach))
		for _, o := range outreach {
			fmt.Printf("    [%s] %s: %s\n", o.CreatedAt, o.ActivityType, o.Notes)
			if o.Outcome != "" {
				fmt.Printf("      Outcome: %s\n", o.Outcome)
			}
		}
	} else {
		fmt.Println("  No activity logged.")
	}
}

func cmdAdd() {
	fmt.Println("\nAdd new lead (press Enter to skip optional fields):")
	company := prompt("Company *")
	if company == "" {
		fmt.Println("Company is required.")
		return
	}
	data := db.Lead{Company: company}
	data.ContactName = prompt("Contact name")
	data.Email = prompt("Email")
	data.Phone = prompt("Phone")
	data.Website = prompt("Website")
	fmt.Println("  Tier: 1=VC  2=Corporate  3=Local  4=Grant  5=Venue  6=Media")
	tier := prompt("Tier [3]")
	if tier == "" {
		tier = "3"
	}
	data.Tier = tier
	data.Type = prompt("Type (VC, Sponsor, Partner...)")
	data.Vertical = prompt("Vertical (fintech, beverage...)")
	data.CheckSize = prompt("Check size")
	data.PitchAngle = prompt("Pitch angle")
	data.NextAction = prompt("Next action")
	data.NextActionDate = prompt("Next action date (YYYY-MM-DD)")
	data.Notes = prompt("Notes")
	id, err := db.AddLead(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	fmt.Printf("\nLead created: %s\n", id)
}

func cmdUpdate() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run crm.go update <id>")
		return
	}
	id := os.Args[2]
	l, err := db.GetLead(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	if l == nil {
		fmt.Println("Lead not found.")
		return
	}
	fmt.Printf("\nEditing: %s (leave blank to keep current value)\n\n", l.Company)
	type field struct {
		name    string
		current string
	}
	fields := []field{
		{"company", l.Company}, {"contact_name", l.ContactName}, {"email", l.Email},
		{"phone", l.Phone}, {"website", l.Website}, {"tier", l.Tier},
		{"type", l.Type}, {"vertical", l.Vertical}, {"check_size", l.CheckSize},
		{"pitch_angle", l.PitchAngle}, {"next_action", l.NextAction},
		{"next_action_date", l.NextActionDate}, {"notes", l.Notes},
	}
	data := map[string]interface{}{}
	for _, f := range fields {
		fmt.Printf("  %s [%s]: ", f.name, f.current)
		var val string
		fmt.Scanln(&val)
		val = strings.TrimSpace(val)
		if val != "" {
			data[f.name] = val
		}
	}
	if len(data) > 0 {
		if err := db.UpdateLead(id, data); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		fmt.Println("Lead updated.")
	} else {
		fmt.Println("No changes.")
	}
}

func cmdDelete() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run crm.go delete <id>")
		return
	}
	id := os.Args[2]
	l, err := db.GetLead(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	if l == nil {
		fmt.Println("Lead not found.")
		return
	}
	fmt.Printf("Delete '%s'? (y/N): ", l.Company)
	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
		if err := db.DeleteLead(id); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		fmt.Println("Deleted.")
	}
}

func cmdStatus() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run crm.go status <id> <new_status>")
		return
	}
	id := os.Args[2]
	status := os.Args[3]
	valid := false
	for _, s := range db.Statuses {
		if s == status {
			valid = true
			break
		}
	}
	if !valid {
		fmt.Printf("Invalid status. Options: %s\n", strings.Join(db.Statuses, ", "))
		return
	}
	if err := db.UpdateLead(id, map[string]interface{}{"status": status}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	fmt.Printf("Status updated to '%s'.\n", status)
}

func cmdLog() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run crm.go log <id>")
		return
	}
	id := os.Args[2]
	l, err := db.GetLead(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	if l == nil {
		fmt.Println("Lead not found.")
		return
	}
	fmt.Printf("\nLog activity for: %s\n", l.Company)
	fmt.Println("Types: email, call, meeting, note")
	activityType := prompt("Type [email]")
	if activityType == "" {
		activityType = "email"
	}
	notes := prompt("Notes")
	outcome := prompt("Outcome")
	if _, err := db.LogOutreach(id, activityType, notes, outcome); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	fmt.Println("Activity logged.")
}

func cmdFollowups() {
	leads, err := db.GetLeads(map[string]string{"status": "active"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	today := time.Now().Format("2006-01-02")
	var due []db.Lead
	for _, l := range leads {
		if l.NextActionDate != "" && l.NextActionDate <= today {
			due = append(due, l)
		}
	}
	if len(due) == 0 {
		fmt.Println("No follow-ups due today.")
		return
	}
	fmt.Printf("\n%d follow-up(s) due:\n\n", len(due))
	for _, l := range due {
		shortID := l.ID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}
		action := ifEmpty(l.NextAction, "No action")
		fmt.Printf("  [%s] %s — %s (%s)\n", shortID, l.Company, action, l.NextActionDate)
	}
}

func cmdImport() {
	path := getFlag("path")
	if path == "" {
		path = "crm-spreadsheet.csv"
	}
	count, err := db.ImportCSV(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	fmt.Printf("Imported %d leads.\n", count)
}

func cmdExport() {
	path := getFlag("path")
	if path == "" {
		path = "leads-export.csv"
	}
	leads, err := db.GetLeads(map[string]string{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	if len(leads) == 0 {
		return
	}
	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	defer f.Close()
	headers := []string{"id", "company", "contact_name", "email", "phone", "website",
		"tier", "type", "vertical", "check_size", "pitch_angle", "status",
		"next_action", "next_action_date", "notes", "source", "created_at", "updated_at"}
	writeCSVLine(f, headers)
	for _, l := range leads {
		writeCSVLine(f, []string{
			l.ID, l.Company, l.ContactName, l.Email, l.Phone, l.Website,
			l.Tier, l.Type, l.Vertical, l.CheckSize, l.PitchAngle, l.Status,
			l.NextAction, l.NextActionDate, l.Notes, l.Source, l.CreatedAt, l.UpdatedAt,
		})
	}
	fmt.Printf("Exported %d leads to %s\n", len(leads), path)
}

func writeCSVLine(f *os.File, vals []string) {
	for i, v := range vals {
		if i > 0 {
			f.WriteString(",")
		}
		if strings.ContainsAny(v, ",\"\n") {
			v = strings.ReplaceAll(v, "\"", "\"\"")
			f.WriteString("\"" + v + "\"")
		} else {
			f.WriteString(v)
		}
	}
	f.WriteString("\n")
}

// ─── HTTP Server (Fiber) ───────────────────────────────────────────────────

func cmdServe() {
	app := fiber.New(fiber.Config{AppName: "WaterParty CRM API"})
	app.Use(cors.New())
	app.Use(logger.New())

	app.Static("/", "./web/build", fiber.Static{
		Index: "index.html",
	})

	api := app.Group("/api")

	api.Get("/stats", func(c *fiber.Ctx) error {
		s, err := db.GetStats()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(s)
	})

	api.Get("/leads", func(c *fiber.Ctx) error {
		filters := map[string]string{}
		if t := c.Query("tier"); t != "" {
			filters["tier"] = t
		}
		if s := c.Query("status"); s != "" {
			filters["status"] = s
		}
		if s := c.Query("search"); s != "" {
			filters["search"] = s
		}
		if v := c.Query("vertical"); v != "" {
			filters["vertical"] = v
		}
		if t := c.Query("type"); t != "" {
			filters["type"] = t
		}
		leads, err := db.GetLeads(filters)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		if leads == nil {
			leads = []db.Lead{}
		}
		return c.JSON(leads)
	})

	api.Get("/leads/:id", func(c *fiber.Ctx) error {
		l, err := db.GetLead(c.Params("id"))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		if l == nil {
			return c.Status(404).JSON(fiber.Map{"error": "lead not found"})
		}
		return c.JSON(l)
	})

	api.Post("/leads", func(c *fiber.Ctx) error {
		var data db.Lead
		if err := c.BodyParser(&data); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		if data.Company == "" {
			return c.Status(400).JSON(fiber.Map{"error": "company is required"})
		}
		id, err := db.AddLead(data)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		data.ID = id
		return c.Status(201).JSON(data)
	})

	api.Put("/leads/:id", func(c *fiber.Ctx) error {
		var data map[string]interface{}
		if err := c.BodyParser(&data); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		if err := db.UpdateLead(c.Params("id"), data); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		l, _ := db.GetLead(c.Params("id"))
		return c.JSON(l)
	})

	api.Delete("/leads/:id", func(c *fiber.Ctx) error {
		if err := db.DeleteLead(c.Params("id")); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"deleted": true})
	})

	api.Put("/leads/:id/status", func(c *fiber.Ctx) error {
		var body struct {
			Status string `json:"status"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		if err := db.UpdateLead(c.Params("id"), map[string]interface{}{"status": body.Status}); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": body.Status})
	})

	api.Get("/leads/:id/outreach", func(c *fiber.Ctx) error {
		outreach, err := db.GetOutreach(c.Params("id"))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		if outreach == nil {
			outreach = []db.OutreachEntry{}
		}
		return c.JSON(outreach)
	})

	api.Post("/leads/:id/outreach", func(c *fiber.Ctx) error {
		var body struct {
			ActivityType string `json:"activity_type"`
			Notes        string `json:"notes"`
			Outcome      string `json:"outcome"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		if body.ActivityType == "" {
			body.ActivityType = "email"
		}
		oid, err := db.LogOutreach(c.Params("id"), body.ActivityType, body.Notes, body.Outcome)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"id": oid})
	})

	api.Post("/leads/import", func(c *fiber.Ctx) error {
		var body struct {
			Path string `json:"path"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		if body.Path == "" {
			body.Path = "crm-spreadsheet.csv"
		}
		count, err := db.ImportCSV(body.Path)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"imported": count})
	})

	api.Get("/followups", func(c *fiber.Ctx) error {
		leads, err := db.GetLeads(map[string]string{"status": "active"})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		today := time.Now().Format("2006-01-02")
		var due []db.Lead
		for _, l := range leads {
			if l.NextActionDate != "" && l.NextActionDate <= today {
				due = append(due, l)
			}
		}
		if due == nil {
			due = []db.Lead{}
		}
		return c.JSON(due)
	})

	app.Use(func(c *fiber.Ctx) error {
		path := c.Path()
		if len(path) < 4 || path[:4] != "/api" {
			c.Set("Content-Type", "text/html; charset=utf-8")
			return c.SendFile("./web/build/index.html")
		}
		return c.Next()
	})

	port := ":8080"
	fmt.Printf("WaterParty CRM API running on http://localhost%s\n", port)
	if err := app.Listen(port); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

// ─── SEND MAIL ─────────────────────────────────────────────────────────────

func cmdMail() {
	emails := getFlag("emails")
	subject := getFlag("subject")
	body := getFlag("body")
	bodyFile := getFlag("body-file")
	fromName := getFlag("from-name")
	dryRun := hasFlag("dry-run")
	confirm := hasFlag("confirm")

	if emails == "" || subject == "" {
		fmt.Println("Usage: go run crm.go mail --emails \"a@b.com,c@d.com\" --subject \"Sub\" --body \"Body\"")
		fmt.Println("  --emails      Comma-separated BCC recipients (required)")
		fmt.Println("  --subject     Email subject (required)")
		fmt.Println("  --body        Email body text")
		fmt.Println("  --body-file   Read body from file")
		fmt.Println("  --attach      File to attach (repeatable)")
		fmt.Println("  --from-name   Sender name (default: John Victor @ WaterParty)")
		fmt.Println("  --dry-run     Preview only")
		fmt.Println("  --confirm     Ask before sending")
		os.Exit(1)
	}

	if body != "" && bodyFile != "" {
		fmt.Println("Use either --body or --body-file, not both")
		os.Exit(1)
	}
	if bodyFile != "" {
		data, err := os.ReadFile(bodyFile)
		if err != nil {
			fmt.Printf("Body file not found: %s\n", bodyFile)
			os.Exit(1)
		}
		body = string(data)
	} else if body == "" && bodyFile == "" {
		fmt.Println("Either --body or --body-file is required")
		os.Exit(1)
	}

	recipients := []string{}
	for _, e := range strings.Split(emails, ",") {
		e = strings.TrimSpace(e)
		if e != "" {
			recipients = append(recipients, e)
		}
	}
	if len(recipients) == 0 {
		fmt.Println("No valid email addresses provided")
		os.Exit(1)
	}

	var invalid []string
	for _, r := range recipients {
		if !validateEmail(r) {
			invalid = append(invalid, r)
		}
	}
	if len(invalid) > 0 {
		fmt.Printf("Invalid email addresses: %s\n", strings.Join(invalid, ", "))
		os.Exit(1)
	}

	attachments := []string{}
	for i, arg := range os.Args {
		if arg == "--attach" && i+1 < len(os.Args) {
			attachments = append(attachments, os.Args[i+1])
		}
	}
	if fromName == "" {
		fromName = "John Victor @ WaterParty"
	}

	if confirm {
		fmt.Printf("\nReady to send to %d recipients via BCC:\n", len(recipients))
		for _, r := range recipients {
			fmt.Printf("   - %s\n", r)
		}
		fmt.Printf("\n   Subject: %s\n", subject)
		for _, a := range attachments {
			fmt.Printf("   Attach:  %s\n", a)
		}
		preview := strings.ReplaceAll(body, "\n", " ")
		if len(preview) > 100 {
			preview = preview[:100]
		}
		fmt.Printf("   Body preview: %s...\n", preview)
		fmt.Print("\n   Send? (y/N): ")
		var confirmStr string
		fmt.Scanln(&confirmStr)
		if strings.ToLower(strings.TrimSpace(confirmStr)) != "y" {
			fmt.Println("Cancelled")
			os.Exit(0)
		}
	}

	sendEmail(recipients, subject, body, fromName, attachments, dryRun)
}

func validateEmail(email string) bool {
	email = strings.TrimSpace(email)
	if strings.Count(email, "@") != 1 {
		return false
	}
	parts := strings.Split(email, "@")
	if parts[0] == "" || parts[1] == "" {
		return false
	}
	if !strings.Contains(parts[1], ".") {
		return false
	}
	return !strings.Contains(email, " ")
}

func loadAppPassword() string {
	conn, err := db.GetMailDB()
	if err != nil {
		fmt.Println("Failed to open credentials DB.")
		fmt.Println("Run: go run crm.go password")
		os.Exit(1)
	}
	defer conn.Close()

	var password string
	err = conn.QueryRow(
		"SELECT app_password FROM credentials WHERE email = ? ORDER BY id DESC LIMIT 1",
		"water.enterprises.org@gmail.com",
	).Scan(&password)
	if err != nil {
		fmt.Println("No credentials found for water.enterprises.org@gmail.com")
		fmt.Println("Run: go run crm.go password")
		os.Exit(1)
	}
	return password
}

func sendEmail(recipients []string, subject, body, fromName string, attachments []string, dryRun bool) int {
	const gmailAddr = "water.enterprises.org@gmail.com"

	if dryRun {
		fmt.Println("\nDRY RUN — No email sent")
		fmt.Printf("   From:       %s <%s>\n", fromName, gmailAddr)
		fmt.Printf("   Subject:    %s\n", subject)
		fmt.Printf("   BCC (%d): %s\n", len(recipients), strings.Join(recipients, ", "))
		for _, a := range attachments {
			if info, err := os.Stat(a); err == nil {
				fmt.Printf("   Attach:     %s (%.1f KB)\n", a, float64(info.Size())/1024)
			}
		}
		fmt.Printf("   Body:\n%s\n\n", body)
		return len(recipients)
	}

	password := loadAppPassword()

	msg := buildMIMEMessage(fromName, subject, body, recipients, attachments)

	auth := smtp.PlainAuth("", gmailAddr, password, "smtp.gmail.com")
	tlsConfig := &tls.Config{ServerName: "smtp.gmail.com"}

	conn, err := tls.Dial("tcp", "smtp.gmail.com:587", tlsConfig)
	if err != nil {
		fmt.Printf("Connection failed: %v\n", err)
		os.Exit(1)
	}

	client, err := smtp.NewClient(conn, "smtp.gmail.com")
	if err != nil {
		fmt.Printf("SMTP client error: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		fmt.Println("SMTP authentication failed.")
		fmt.Println("Run: go run crm.go password")
		os.Exit(1)
	}

	if err := client.Mail(gmailAddr); err != nil {
		fmt.Printf("Mail from error: %v\n", err)
		os.Exit(1)
	}
	for _, r := range recipients {
		if err := client.Rcpt(r); err != nil {
			fmt.Printf("Recipient refused %s: %v\n", r, err)
			os.Exit(1)
		}
	}

	w, err := client.Data()
	if err != nil {
		fmt.Printf("Data error: %v\n", err)
		os.Exit(1)
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		fmt.Printf("Write error: %v\n", err)
		os.Exit(1)
	}
	w.Close()

	attachInfo := ""
	if len(attachments) > 0 {
		attachInfo = fmt.Sprintf(" with %d attachment(s)", len(attachments))
	}
	fmt.Printf("Email sent to %d recipients via BCC%s\n", len(recipients), attachInfo)
	fmt.Printf("   Subject: %s\n", subject)
	return len(recipients)
}

func buildMIMEMessage(fromName, subject, body string, recipients, attachments []string) string {
	const gmailAddr = "water.enterprises.org@gmail.com"
	var buf strings.Builder
	boundary := "waterparty-boundary-123"
	altBoundary := "waterparty-alt-boundary-456"

	buf.WriteString(fmt.Sprintf("From: %s <%s>\r\n", mime.QEncoding.Encode("utf-8", fromName), gmailAddr))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", gmailAddr))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.QEncoding.Encode("utf-8", subject)))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
	buf.WriteString("\r\n")
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", altBoundary))
	buf.WriteString("\r\n")
	buf.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
	buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	buf.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(body)
	buf.WriteString("\r\n")
	buf.WriteString(fmt.Sprintf("--%s--\r\n", altBoundary))
	buf.WriteString("\r\n")

	for _, apath := range attachments {
		data, err := os.ReadFile(apath)
		if err != nil {
			fmt.Printf("Attachment not found: %s\n", apath)
			os.Exit(1)
		}
		_, fname := filepath.Split(apath)
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString("Content-Type: application/octet-stream\r\n")
		buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", fname))
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString("\r\n")
		encoded := encodeBase64(data)
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			buf.WriteString(encoded[i:end])
			buf.WriteString("\r\n")
		}
		buf.WriteString("\r\n")
	}

	buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	return buf.String()
}

func encodeBase64(data []byte) string {
	const table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var result strings.Builder
	for i := 0; i < len(data); i += 3 {
		var b24 uint32
		remaining := len(data) - i
		b := make([]byte, 4)
		b24 = uint32(data[i]) << 16
		if remaining > 1 {
			b24 |= uint32(data[i+1]) << 8
		}
		if remaining > 2 {
			b24 |= uint32(data[i+2])
		}
		b[0] = table[(b24>>18)&0x3F]
		b[1] = table[(b24>>12)&0x3F]
		if remaining > 1 {
			b[2] = table[(b24>>6)&0x3F]
		} else {
			b[2] = '='
		}
		if remaining > 2 {
			b[3] = table[b24&0x3F]
		} else {
			b[3] = '='
		}
		result.Write(b)
	}
	return result.String()
}

// ─── STORE PASSWORD ────────────────────────────────────────────────────────

func cmdPassword() {
	password := ""
	for i, arg := range os.Args {
		if arg == "--password" && i+1 < len(os.Args) {
			password = os.Args[i+1]
		}
		if strings.HasPrefix(arg, "--password=") {
			password = strings.SplitN(arg, "=", 2)[1]
		}
	}

	fmt.Println("WaterParty — Store Gmail App Password")
	fmt.Printf("   Account: %s\n", "water.enterprises.org@gmail.com")
	fmt.Println("   Database: SQLCipher AES-256 encrypted")
	fmt.Println()

	if password != "" {
		fmt.Println("Warning: --password is visible in process listings and shell history.")
		fmt.Println()
	} else {
		pw1 := readPasswordSecure("Enter Gmail app password: ")
		if pw1 == "" {
			fmt.Println("Password cannot be empty")
			os.Exit(1)
		}
		pw2 := readPasswordSecure("Confirm password: ")
		if pw1 != pw2 {
			fmt.Println("Passwords do not match")
			os.Exit(1)
		}
		password = pw1
	}

	storePassword("water.enterprises.org@gmail.com", password)
}

func readPasswordSecure(promptMsg string) string {
	fmt.Print(promptMsg)
	raw, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		fmt.Println("Failed to read password")
		os.Exit(1)
	}
	return strings.TrimSpace(string(raw))
}

func storePassword(email, appPassword string) {
	if _, err := os.Stat(db.ENVPath); os.IsNotExist(err) {
		fmt.Printf(".env file not found at %s\n", db.ENVPath)
		fmt.Println("   Create it and add: EMAIL_DB_PASSWORD=\"your_password\"")
		os.Exit(1)
	}

	pw := db.LoadDBPassword()
	hexKey := fmt.Sprintf("%x", []byte(pw))
	dsn := fmt.Sprintf("%s?_pragma_key=x'%s'", db.MailDBPath, hexKey)

	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	conn.Exec("PRAGMA journal_mode=WAL")
	conn.Exec(`CREATE TABLE IF NOT EXISTS credentials (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		app_password TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`)

	_, err = conn.Exec("INSERT OR REPLACE INTO credentials (email, app_password) VALUES (?, ?)", email, appPassword)
	if err != nil {
		fmt.Printf("Failed to store password: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("App password stored in SQLCipher-encrypted database: %s\n", db.MailDBPath)
	fmt.Println("   Database encryption password: EMAIL_DB_PASSWORD from .env")
	fmt.Println("   App password stored in plain text inside the encrypted DB")
	fmt.Println()
	fmt.Println("You can now use: go run crm.go mail")
	fmt.Println("   --emails \"addr1@example.com,addr2@example.com\"")
	fmt.Println("   --subject \"Your Subject\"")
	fmt.Println("   --body \"Your email body text\"")
}

// ─── SEND TO TELEGRAM ──────────────────────────────────────────────────────

func cmdTelegram() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if token == "" {
		fmt.Println("Erro: TELEGRAM_BOT_TOKEN nao definido.")
		fmt.Println("      Adicione ao arquivo .env ou exporte como variavel de ambiente.")
		os.Exit(1)
	}
	if chatID == "" {
		fmt.Println("Erro: TELEGRAM_CHAT_ID nao definido.")
		fmt.Println("      Adicione ao arquivo .env ou exporte como variavel de ambiente.")
		os.Exit(1)
	}

	contactList := "contact@monashees.com, contato@domo.vc, info@canary.com.br, biz@cesar.org.br"

	emailMonashees := `EMAIL 1 — Monashees (VC - English)
To: contact@monashees.com
Subject: WaterParty — Tinder for parties, launching in Recife

Hi Monashees team,

I'm reaching out from Water Enterprises (Stellarium Foundation). We built WaterParty — a Tinder-style app for discovering parties and events, with integrated payments (tipping + crowdfunding) and auto-currency detection via GPS.

Why it matters: The global nightlife market is $150B+. No app combines discovery + chat + payments in one place. We do.

Traction: Production-ready MVP (React 19, Bun, Turso, Stripe). Cross-platform (iOS + Android). Multi-currency GPS detection. WebSocket real-time.

Launch strategy: Recife first (4M pop, 100K+ students, Porto Digital). Prove density, expand city-by-city.

The ask: Raising $250K-$500K pre-seed to acquire our first 25K users and expand to new cities.

Best,
John Victor
Water Enterprises / Stellarium Foundation
water.enterprises.org@gmail.com`

	emailDomo := `EMAIL 2 — DOMO.VC (VC - English)
To: contato@domo.vc
Subject: WaterParty — Tinder for parties, launching in Recife

Hi DOMO team,

I'm reaching out from Water Enterprises. We built WaterParty — a Tinder-style app for discovering parties and events, with integrated payments and auto-currency detection.

The ask: Raising $250K-$500K pre-seed launching in Recife.

Best,
John Victor
water.enterprises.org@gmail.com`

	emailCesar := `EMAIL 3 — CESAR Recife (Local - Portugues)
To: biz@cesar.org.br
Assunto: WaterParty — App de descoberta de festas, lancando em Recife

Ola equipe CESAR,

Sou da Water Enterprises e estamos lancando o WaterParty em Recife — um app estilo Tinder para descobrir festas e eventos, com pagamentos integrados e deteccao automatica de moeda por GPS.

Tracao: MVP pronto (React 19, Bun, Turso, Stripe). iOS + Android.

Pedido: Captando R$ 1M-R$ 2M pre-seed.

Gostaria de saber mais sobre programas de incubacao com o CESAR Labs.

Abraco,
John Victor
water.enterprises.org@gmail.com`

	message := fmt.Sprintf(`WATERPARTY EMAIL CAMPAIGN

CONTATOS PARA ENVIAR EMAIL:
%s

---

%s

---

%s

---

%s`, contactList, emailMonashees, emailDomo, emailCesar)

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := fmt.Sprintf(`{"chat_id":%s,"text":%s}`, chatID, jsonEscape(message))

	fmt.Println("Enviando campanha de email do WaterParty para o Telegram...")
	fmt.Printf("  Chat ID: %s\n", chatID)
	fmt.Printf("  Tamanho da mensagem: %d caracteres\n", len(message))
	fmt.Println()

	resp, err := http.Post(url, "application/json", strings.NewReader(payload))
	if err != nil {
		fmt.Printf("Erro de conexao: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		fmt.Println("OK! Campanha enviada com sucesso!")
	} else {
		fmt.Printf("Falha ao enviar (HTTP %d):\n%s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}
}

func jsonEscape(s string) string {
	result := ""
	for _, r := range s {
		switch r {
		case '\\':
			result += "\\\\"
		case '"':
			result += "\\\""
		case '\n':
			result += "\\n"
		case '\r':
			result += "\\r"
		case '\t':
			result += "\\t"
		default:
			result += string(r)
		}
	}
	return "\"" + result + "\""
}

// ─── COLD CAMPAIGN ─────────────────────────────────────────────────────────

type campaign struct {
	Lid     string
	Email   string
	Subject string
	Body    string
}

func cmdCampaign() {
	send := hasFlag("send")
	dryRun := hasFlag("dry-run")

	campaigns := buildCampaigns()

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  WATERPARTY — COLD LEAD EMAIL CAMPAIGN")
	fmt.Printf("  Total: %d personalized emails\n", len(campaigns))
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\nCAMPAIGN PREVIEW:\n")

	for i, c := range campaigns {
		subjPreview := c.Subject
		if len(subjPreview) > 70 {
			subjPreview = subjPreview[:70]
		}
		fmt.Printf("  %2d. [%s] %s...\n", i+1, c.Email, subjPreview)
	}

	if !send && !dryRun {
		fmt.Println("\n  Use --dry-run to preview or --send to send")
		return
	}

	if dryRun {
		fmt.Println("\n  DRY RUN — no emails sent")
		for i, c := range campaigns {
			fmt.Printf("\n  [%d/%d] To: %s\n", i+1, len(campaigns), c.Email)
			fmt.Printf("       Subject: %s\n", c.Subject)
			preview := c.Body
			if len(preview) > 100 {
				preview = preview[:100]
			}
			fmt.Printf("       Body preview: %s...\n", preview)
		}
		return
	}

	if send {
		sent := 0
		errors := 0
		for i, c := range campaigns {
			fmt.Printf("  [%d/%d] Sending to %s... ", i+1, len(campaigns), c.Email)
			ok, output := sendOne(c.Email, c.Subject, c.Body)
			if ok {
				db.LogOutreach(c.Lid, "email", "Cold campaign outreach",
					fmt.Sprintf("Sent to %s", c.Email))
				db.UpdateLead(c.Lid, map[string]interface{}{
					"status":           "contacted",
					"next_action":      "Monitorar reply",
					"next_action_date": "2026-06-27",
				})
				sent++
				fmt.Println("OK")
			} else {
				errors++
				errMsg := output
				if len(errMsg) > 200 {
					errMsg = errMsg[:200]
				}
				fmt.Printf("FAIL: %s\n", errMsg)
			}
		}

		fmt.Printf("\n%s\n", strings.Repeat("=", 60))
		fmt.Printf("  CAMPAIGN COMPLETE\n")
		fmt.Printf("  Sent: %d | Errors: %d\n", sent, errors)
		fmt.Println(strings.Repeat("=", 60))

		remaining := []string{"Influency.me", "HypeAuditor", "Lessie AI",
			"TikTok Creator Marketplace", "Instagram Creator Marketplace",
			"100 Open Startups", "LAVCA", "Made Assessoria"}
		fmt.Println("\n  Remaining cold leads (need platform signup, not email):")
		for _, name := range remaining {
			fmt.Printf("    - %s\n", name)
		}
	}
}

func sendOne(email, subject, body string) (bool, string) {
	oldArgs := os.Args
	os.Args = []string{"crm", "mail",
		"--emails", email,
		"--subject", subject,
		"--body", body,
	}

	cmdMail()

	os.Args = oldArgs
	fmt.Println("---")
	return true, ""
}

func buildCampaigns() []campaign {
	appURL := "https://waterparty-react-14hr.onrender.com"
	landingURL := "https://water-enterprises-landing.onrender.com"
	githubURL := "https://github.com/StellariumFoundation/WaterParty-React"

	return []campaign{
		{
			Lid: "7ea1b4dd-2826-45ca-bf1c-3419bcf240c2", Email: "marketing@ambev.com.br",
			Subject: "Parceria WaterParty x AMBEV: Patrocínio Categoria Cerveja no App",
			Body:    buildAmbevBody(appURL, landingURL, githubURL),
		},
		{
			Lid: "d72c1210-0042-49d9-af7b-70464cb398d2", Email: "portodigital@portodigital.org",
			Subject: "WaterParty — Candidatura Programa de Incubação Porto Digital",
			Body:    fmt.Sprintf("Prezados, Porto Digital,\n\nMeu nome é John Victor, founder do WaterParty (Water Enterprises). Somos uma startup de tecnologia sediada em Recife construindo o maior app de descoberta de festas e eventos do Brasil.\n\nLinks:\n- App: %s\n- Landing: %s\n- GitHub: %s\n\nASK: Gostaria de candidatar o WaterParty aos programas de incubação.\n\nAtenciosamente,\nJohn Victor\nwater.enterprises.org@gmail.com", appURL, landingURL, githubURL),
		},
		{
			Lid: "472b0c49-6919-47cc-8f1b-2b058aa5fe6b", Email: "facepe@facepe.br",
			Subject: "WaterParty — Consulta Editais Inovação FACEPE (Startup Recife/PE)",
			Body:    fmt.Sprintf("Prezados, FACEPE,\n\nWaterParty é uma startup pernambucana de tecnologia.\n\nLinks: %s | %s\n\nASK: Há editais abertos para 2026/2027?\n\nJohn Victor\nwater.enterprises.org@gmail.com", appURL, landingURL),
		},
		{
			Lid: "d89d96d0-57be-44e7-96ed-5d75e46bf7de", Email: "chamada-tecnova@fapesp.br",
			Subject: "WaterParty — Interesse Programa Tecnova 2026/2027",
			Body:    fmt.Sprintf("Prezados,\n\nWaterParty — startup brasileira de tecnologia.\n\nLinks: %s | %s | %s\n\nASK: Gostaria de orientações sobre o Tecnova 2026/2027.\n\nAtenciosamente,\nJohn Victor\nwater.enterprises.org@gmail.com\n%s", appURL, landingURL, githubURL, appURL),
		},
		{
			Lid: "f05438ed-8f28-408d-bb23-5cf7e169c6ae", Email: "bndesgaragem@quintessa.org.br",
			Subject: "WaterParty — Candidatura BNDES Garagem (Negócio de Impacto)",
			Body:    fmt.Sprintf("Prezados, BNDES Garagem,\n\nWaterParty conecta jovens a experiências reais.\n\nLinks: %s | %s\n\nASK: Gostaria de candidatar ao BNDES Garagem.\n\nJohn Victor\nwater.enterprises.org@gmail.com", appURL, landingURL),
		},
		{
			Lid: "466a102c-d757-48e5-ad74-c0ac3aec86b5", Email: "directoria@cnpq.br",
			Subject: "WaterParty — Consulta Programa RHAE / Bolsas Inovação CNPq",
			Body:    fmt.Sprintf("Prezados, CNPq,\n\nWaterParty é uma startup pernambucana.\n\nLinks: %s | %s\n\nASK: Há editais abertos do Programa RHAE?\n\nJohn Victor\nwater.enterprises.org@gmail.com", appURL, landingURL),
		},
		{
			Lid: "b0ecdf05-a677-4b18-bd88-1c7092551866", Email: "sebraepe@sebrae.com.br",
			Subject: "WaterParty — Consulta Programas SEBRAE Startups PE",
			Body:    fmt.Sprintf("Prezados, SEBRAE PE,\n\nWaterParty é uma startup pernambucana de tecnologia.\n\nLinks: %s | %s\n\nASK: Há programas para startups?\n\nJohn Victor\nwater.enterprises.org@gmail.com", appURL, landingURL),
		},
		{
			Lid: "fa92cc68-d9f9-45e8-8594-1bf46d2ddaf9", Email: "googleforstartups@google.com",
			Subject: "WaterParty — Application Google for Startups Brazil",
			Body:    fmt.Sprintf("Dear Google for Startups Team,\n\nWaterParty — social discovery platform.\n\nLinks:\n- Live app: %s\n- Landing: %s\n- GitHub: %s\n\nASK: Applying for Google for Startups Brazil.\n\nBest,\nJohn Victor\nwater.enterprises.org@gmail.com", appURL, landingURL, githubURL),
		},
		{
			Lid: "908b0b11-697a-40fb-88c5-698a87714b32", Email: "hatab@4equity.com.br",
			Subject: "WaterParty — Proposta Media for Equity",
			Body:    fmt.Sprintf("Olá Felipe,\n\nWaterParty — Media for Equity.\n\nLinks: %s | %s | %s\n\nASK: Vamos conversar?\n\nJohn Victor\nwater.enterprises.org@gmail.com\n%s", appURL, landingURL, githubURL, appURL),
		},
		{
			Lid: "ef70d1dc-12eb-4073-bf1b-b211df0c0190", Email: "contato@nexpon.com.br",
			Subject: "WaterParty — Proposta Media for Equity (Expansão NE Brasil)",
			Body:    fmt.Sprintf("Olá, equipe Nexpon,\n\nWaterParty — Media for Equity.\n\nLinks: %s | %s\n\nASK: Vamos conversar?\n\nJohn Victor\nwater.enterprises.org@gmail.com", appURL, landingURL),
		},
		{
			Lid: "e5816bf7-40d3-439b-8b1f-07e0953c73c7", Email: "halisson@boldcomunicacao.com.br",
			Subject: "Parceria WaterParty x Bold Comunicação",
			Body:    fmt.Sprintf("Olá Halisson,\n\nWaterParty — app pernambucano.\n\nLinks: %s | %s\n\nASK: Interesse em parceria?\n\nJohn Victor\nwater.enterprises.org@gmail.com", appURL, landingURL),
		},
		{
			Lid: "58a53400-d709-4236-9595-478b3f75e5a1", Email: "contato@agenciacosmica.com.br",
			Subject: "Parceria WaterParty x Agência Cósmica",
			Body:    fmt.Sprintf("Olá, equipe Cósmica,\n\nWaterParty — app pernambucano.\n\nLinks: %s | %s\n\nASK: Interesse em parceria?\n\nJohn Victor\nwater.enterprises.org@gmail.com", appURL, landingURL),
		},
		{
			Lid: "b3f3985d-4d8d-4f22-a2c9-126bb6f6f911", Email: "contato@vdbconecta.com.br",
			Subject: "Parceria WaterParty x VDB Conecta — Influenciadores NE",
			Body:    fmt.Sprintf("Olá, equipe VDB Conecta,\n\nWaterParty — app de descoberta de festas.\n\nLinks: %s | %s\n\nASK: Parceria de influenciadores?\n\nJohn Victor\nwater.enterprises.org@gmail.com", appURL, landingURL),
		},
	}
}

func buildAmbevBody(appURL, landingURL, githubURL string) string {
	return fmt.Sprintf(`Olá, equipe AMBEV!

Meu nome é John Victor, founder do WaterParty — um aplicativo de descoberta social de festas e eventos.

Links:
- App: %s
- Landing: %s
- GitHub: %s

OPORTUNIDADE DE PATROCÍNIO: Categoria Exclusiva de Cerveja no WaterParty.

ASK: Gostaria de agendar uma call de 15 min.

Atenciosamente,
John Victor
water.enterprises.org@gmail.com
%s`, appURL, landingURL, githubURL, appURL)
}
