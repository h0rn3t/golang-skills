---
description: 'Extraction request on existing code; the code is pasted because each run starts in an empty directory'
tags: [trigger]
runs: 3
max_turns: 10
allowed_tools: [Read, Glob, Grep, Skill]
---

Split this function into smaller ones:

```go
func HandleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	f, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer f.Close()
	rd := csv.NewReader(f)
	rows, err := rd.ReadAll()
	if err != nil {
		http.Error(w, "bad csv", http.StatusBadRequest)
		return
	}
	var users []User
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) != 3 {
			http.Error(w, fmt.Sprintf("row %d: want 3 fields", i), http.StatusBadRequest)
			return
		}
		age, err := strconv.Atoi(row[2])
		if err != nil || age < 0 {
			http.Error(w, fmt.Sprintf("row %d: bad age", i), http.StatusBadRequest)
			return
		}
		if !strings.Contains(row[1], "@") {
			http.Error(w, fmt.Sprintf("row %d: bad email", i), http.StatusBadRequest)
			return
		}
		users = append(users, User{Name: row[0], Email: row[1], Age: age})
	}
	tx, err := db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "db", http.StatusInternalServerError)
		return
	}
	for _, u := range users {
		if _, err := tx.ExecContext(r.Context(),
			"INSERT INTO users(name, email, age) VALUES ($1, $2, $3)", u.Name, u.Email, u.Age); err != nil {
			tx.Rollback()
			http.Error(w, "db", http.StatusInternalServerError)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, "db", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "imported %d users\n", len(users))
}
```
