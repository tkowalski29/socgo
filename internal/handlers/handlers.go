package handlers

import (
	"net/http"
	"html/template"
	"log"
)

var homeTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SocGo</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <h1 class="text-4xl font-bold text-center text-gray-800 mb-8">
            Welcome to SocGo
        </h1>
        <div class="max-w-md mx-auto bg-white rounded-lg shadow-md p-6">
            <p class="text-gray-600 text-center">
                A simple social media app built with Go, HTMX, and Tailwind CSS.
            </p>
            <button 
                hx-get="/health" 
                hx-target="#status"
                class="mt-4 w-full bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
            >
                Check Status
            </button>
            <div id="status" class="mt-4 text-center"></div>
        </div>
    </div>
</body>
</html>
`

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("home").Parse(homeTemplate)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<div class="p-2 bg-green-100 text-green-800 rounded">✓ Server is healthy</div>`))
}