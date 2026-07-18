package report

import (
	"html/template"
	"os"
	"sort"

	"takeout_services/internal/domain/model"
)

// GenerateHTMLReport outputs a visual dashboard report to targetPath.
func GenerateHTMLReport(services []*model.DetectedService, targetPath string) error {
	// Sort by confidence (descending) then by name
	sort.Slice(services, func(i, j int) bool {
		if services[i].Confidence == services[j].Confidence {
			return services[i].Name < services[j].Name
		}
		return services[i].Confidence > services[j].Confidence
	})

	t, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Calculate overall stats
	stats := struct {
		Total        int
		HighConf     int
		Subscriptions int
	}{
		Total: len(services),
	}

	for _, s := range services {
		if s.Confidence >= 7 {
			stats.HighConf++
		}
		if s.HasReceipt {
			stats.Subscriptions++
		}
	}

	data := struct {
		Services []*model.DetectedService
		Stats    interface{}
	}{
		Services: services,
		Stats:    stats,
	}

	return t.Execute(f, data)
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta name="description" content="Digital footprint analyzer report showing web service registrations, payment receipts, and password resets from email history with direct deletion links.">
	<title>Digital Footprint Analyzer</title>
	<link rel="preconnect" href="https://fonts.googleapis.com">
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
	<link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;500;600;700&display=swap" rel="stylesheet">
	<style>
		:root {
			--bg-dark: #F8FAFC;
			--panel-dark: #FFFFFF;
			--accent: #0284C7;
			--accent-glow: rgba(2, 132, 199, 0.1);
			--text-main: #0F172A;
			--text-muted: #334155; /* Darker grey Slate-700 for 7:1 contrast ratio against white */
			--border: #CBD5E1; /* Slate-300 for distinct card borders */

			--welcome-color: #0369A1; /* Sky-700 for 5.9:1 contrast ratio */
			--welcome-bg: rgba(3, 105, 161, 0.08);
			--welcome-border: rgba(3, 105, 161, 0.2);

			--reset-color: #B45309;
			--reset-bg: rgba(245, 158, 11, 0.08);
			--reset-border: rgba(245, 158, 11, 0.25);

			--receipt-color: #0F172A;
			--receipt-bg: rgba(15, 23, 42, 0.05);
			--receipt-border: rgba(15, 23, 42, 0.12);

			--unclassified-bg: rgba(100, 116, 139, 0.06);
			--unclassified-border: rgba(100, 116, 139, 0.15);

			--count-bg: rgba(100, 116, 139, 0.06);
			--count-border: rgba(100, 116, 139, 0.15);

			--conf-high-color: #0369A1;
			--conf-high-bg: rgba(2, 132, 199, 0.08);
			--conf-high-border: rgba(2, 132, 199, 0.2);

			--conf-mid-color: #B45309;
			--conf-mid-bg: rgba(245, 158, 11, 0.08);
			--conf-mid-border: rgba(245, 158, 11, 0.25);

			--conf-low-color: #DC2626;
			--conf-low-bg: rgba(220, 30, 30, 0.08);
			--conf-low-border: rgba(220, 30, 30, 0.2);

			--delete-btn-bg: rgba(220, 38, 38, 0.08);
			--delete-btn-text: #DC2626;
			--delete-btn-border: rgba(220, 38, 38, 0.2);
			--delete-btn-hover-bg: #DC2626;
			--delete-btn-hover-border: #DC2626;
		}

		body.dark-mode {
			--bg-dark: #1B0C0C;
			--panel-dark: #2E1615;
			--accent: #90DD55;
			--accent-glow: rgba(144, 221, 85, 0.15);
			--text-main: #F2E4E3;
			--text-muted: #C8A679;
			--border: rgba(200, 166, 121, 0.15);

			--welcome-color: #90DD55;
			--welcome-bg: rgba(144, 221, 85, 0.1);
			--welcome-border: rgba(144, 221, 85, 0.2);

			--reset-color: #C8A679;
			--reset-bg: rgba(200, 166, 121, 0.1);
			--reset-border: rgba(200, 166, 121, 0.2);

			--receipt-color: #F2E4E3;
			--receipt-bg: rgba(242, 228, 227, 0.1);
			--receipt-border: rgba(242, 228, 227, 0.2);

			--unclassified-bg: rgba(242, 228, 227, 0.05);
			--unclassified-border: rgba(242, 228, 227, 0.15);

			--count-bg: rgba(200, 166, 121, 0.1);
			--count-border: rgba(200, 166, 121, 0.2);

			--conf-high-color: #90DD55;
			--conf-high-bg: rgba(144, 221, 85, 0.15);
			--conf-high-border: rgba(144, 221, 85, 0.3);

			--conf-mid-color: #C8A679;
			--conf-mid-bg: rgba(200, 166, 121, 0.15);
			--conf-mid-border: rgba(200, 166, 121, 0.3);

			--conf-low-color: #F2E4E3;
			--conf-low-bg: rgba(242, 228, 227, 0.1);
			--conf-low-border: rgba(242, 228, 227, 0.2);

			--delete-btn-bg: rgba(239, 68, 68, 0.1);
			--delete-btn-text: #ef4444;
			--delete-btn-border: rgba(239, 68, 68, 0.2);
			--delete-btn-hover-bg: #ef4444;
			--delete-btn-hover-border: #ef4444;
		}

		body.dark-mode h1 {
			background: linear-gradient(135deg, #F2E4E3, #90DD55, #C8A679);
			-webkit-background-clip: text;
			-webkit-text-fill-color: transparent;
		}

		body.dark-mode .stat-icon {
			opacity: 0.25;
		}

		.theme-toggle-wrapper {
			position: absolute;
			top: 2rem;
			right: 1.5rem;
		}
		
		@media (max-width: 768px) {
			.theme-toggle-wrapper {
				position: static;
				margin-bottom: 1.5rem;
				text-align: center;
				display: flex;
				justify-content: center;
			}
		}

		* {
			box-sizing: border-box;
			margin: 0;
			padding: 0;
		}

		body {
			font-family: 'Outfit', sans-serif;
			background-color: var(--bg-dark);
			color: var(--text-main);
			min-height: 100vh;
			padding: 2.5rem 1.5rem;
			line-height: 1.5;
		}

		.container {
			max-width: 1200px;
			margin: 0 auto;
		}

		header {
			margin-bottom: 2.5rem;
			text-align: center;
		}

		h1 {
			font-size: 2.5rem;
			font-weight: 700;
			background: linear-gradient(135deg, #0F172A, #0284C7, #64748B);
			-webkit-background-clip: text;
			-webkit-text-fill-color: transparent;
			margin-bottom: 0.5rem;
			letter-spacing: -0.025em;
		}

		p.subtitle {
			color: var(--text-muted);
			font-size: 1.1rem;
		}

		/* Stats Grid */
		.stats-grid {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
			gap: 1.5rem;
			margin-bottom: 3rem;
		}

		.stat-card {
			background-color: var(--panel-dark);
			border: 1px solid var(--border);
			border-radius: 1rem;
			padding: 1.5rem;
			display: flex;
			align-items: center;
			justify-content: space-between;
			position: relative;
			overflow: hidden;
			transition: border-color 0.3s;
		}

		.stat-card:hover {
			border-color: var(--accent);
		}

		.stat-info h2 {
			font-size: 0.9rem;
			color: var(--text-muted);
			text-transform: uppercase;
			letter-spacing: 0.05em;
			margin-bottom: 0.25rem;
		}

		.stat-value {
			font-size: 2.2rem;
			font-weight: 700;
			color: var(--text-main);
		}

		.stat-icon {
			font-size: 2.5rem;
			opacity: 0.12;
			color: var(--accent);
		}

		/* Search & Filter Bar */
		.controls-bar {
			background-color: var(--panel-dark);
			border: 1px solid var(--border);
			border-radius: 1rem;
			padding: 1.25rem;
			margin-bottom: 2rem;
			display: flex;
			flex-wrap: wrap;
			gap: 1rem;
			align-items: center;
			justify-content: space-between;
		}

		.search-wrapper {
			position: relative;
			flex: 1;
			min-width: 280px;
		}

		.search-input {
			width: 100%;
			background-color: var(--bg-dark);
			border: 1px solid var(--border);
			color: var(--text-main);
			padding: 0.75rem 1rem;
			border-radius: 0.75rem;
			font-family: inherit;
			font-size: 0.95rem;
			outline: none;
			transition: border-color 0.2s, box-shadow 0.2s;
		}

		.search-input:focus {
			border-color: var(--accent);
			box-shadow: 0 0 0 3px var(--accent-glow);
		}

		.filters {
			display: flex;
			gap: 0.5rem;
		}

		.filter-btn {
			background-color: var(--bg-dark);
			border: 1px solid var(--border);
			color: var(--text-muted);
			padding: 0.5rem 1rem;
			border-radius: 0.5rem;
			font-family: inherit;
			font-weight: 500;
			font-size: 0.9rem;
			cursor: pointer;
			transition: all 0.2s;
		}

		.filter-btn:hover, .filter-btn.active {
			background-color: var(--accent);
			border-color: var(--accent);
			color: var(--text-main);
		}

		/* Services Grid */
		.services-grid {
			display: grid;
			grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
			gap: 1.5rem;
		}

		.service-card {
			background-color: var(--panel-dark);
			border: 1px solid var(--border);
			border-radius: 1rem;
			padding: 1.5rem;
			display: flex;
			flex-direction: column;
			position: relative;
			transition: transform 0.2s, border-color 0.2s, box-shadow 0.2s;
		}

		.service-card:hover {
			transform: translateY(-2px);
			border-color: var(--accent);
			box-shadow: 0 8px 30px rgba(0, 0, 0, 0.4), 0 0 15px var(--accent-glow);
		}

		.card-header {
			display: flex;
			justify-content: space-between;
			align-items: flex-start;
			margin-bottom: 1rem;
		}

		.service-identity h3 {
			font-size: 1.3rem;
			font-weight: 600;
			color: var(--text-main);
			margin-bottom: 0.1rem;
		}

		.service-domain {
			font-size: 0.85rem;
			color: var(--text-muted);
		}

		.confidence-badge {
			padding: 0.25rem 0.6rem;
			border-radius: 0.5rem;
			font-size: 0.75rem;
			font-weight: 600;
			text-transform: uppercase;
		}

		.conf-high { background-color: var(--conf-high-bg); color: var(--conf-high-color); border: 1px solid var(--conf-high-border); }
		.conf-mid { background-color: var(--conf-mid-bg); color: var(--conf-mid-color); border: 1px solid var(--conf-mid-border); }
		.conf-low { background-color: var(--conf-low-bg); color: var(--conf-low-color); border: 1px solid var(--conf-low-border); }

		/* Indicator Badges */
		.indicators {
			display: flex;
			flex-wrap: wrap;
			gap: 0.5rem;
			margin-bottom: 1.25rem;
		}

		.badge {
			font-size: 0.75rem;
			padding: 0.2rem 0.5rem;
			border-radius: 0.35rem;
			font-weight: 500;
		}

		.badge.welcome { background-color: var(--welcome-bg); color: var(--welcome-color); border: 1px solid var(--welcome-border); }
		.badge.reset { background-color: var(--reset-bg); color: var(--reset-color); border: 1px solid var(--reset-border); }
		.badge.receipt { background-color: var(--receipt-bg); color: var(--receipt-color); border: 1px solid var(--receipt-border); }
		.badge.unclassified { background-color: var(--unclassified-bg); color: var(--text-muted); border: 1px solid var(--unclassified-border); }
		.badge.count { background-color: var(--count-bg); color: var(--text-muted); border: 1px solid var(--count-border); }

		/* Expandable Verification Info */
		.verify-section {
			margin-top: auto;
			padding-top: 1rem;
			border-top: 1px solid var(--border);
		}

		.toggle-subjects {
			background: none;
			border: none;
			color: var(--accent);
			font-family: inherit;
			font-size: 0.85rem;
			font-weight: 500;
			cursor: pointer;
			display: flex;
			align-items: center;
			gap: 0.25rem;
			outline: none;
		}

		.toggle-subjects::after {
			content: '▼';
			font-size: 0.65rem;
			transition: transform 0.2s;
		}

		.toggle-subjects.active::after {
			transform: rotate(180deg);
		}

		.subjects-list {
			display: none;
			margin-top: 0.5rem;
			padding: 0.5rem;
			background-color: var(--bg-dark);
			border-radius: 0.5rem;
			font-size: 0.8rem;
			color: var(--text-muted);
			border: 1px solid var(--border);
		}

		.subjects-list.show {
			display: block;
		}

		.subjects-list li {
			margin-bottom: 0.35rem;
			list-style-type: none;
			padding-left: 0.5rem;
			border-left: 2px solid var(--accent);
			overflow: hidden;
			text-overflow: ellipsis;
			white-space: nowrap;
		}

		.subjects-list li:last-child {
			margin-bottom: 0;
		}

		/* Custom Checkbox Styling */
		.deleted-checkbox-wrapper {
			display: inline-block;
			position: relative;
			width: 1.3rem;
			height: 1.3rem;
			margin-top: 0.15rem;
			cursor: pointer;
			user-select: none;
		}

		.deleted-checkbox-wrapper input {
			position: absolute;
			opacity: 0;
			cursor: pointer;
			height: 0;
			width: 0;
		}

		.checkmark {
			position: absolute;
			top: 0;
			left: 0;
			height: 1.3rem;
			width: 1.3rem;
			background-color: var(--panel-dark);
			border: 2px solid var(--border);
			border-radius: 0.35rem;
			transition: all 0.2s;
		}

		.deleted-checkbox-wrapper:hover input ~ .checkmark {
			border-color: var(--accent);
		}

		.deleted-checkbox-wrapper input:checked ~ .checkmark {
			background-color: var(--accent);
			border-color: var(--accent);
		}

		.checkmark:after {
			content: "";
			position: absolute;
			display: none;
		}

		.deleted-checkbox-wrapper input:checked ~ .checkmark:after {
			display: block;
		}

		.deleted-checkbox-wrapper .checkmark:after {
			left: 5px;
			top: 1px;
			width: 4px;
			height: 8px;
			border: solid #ffffff;
			border-width: 0 2px 2px 0;
			transform: rotate(45deg);
		}

		/* Deleted Card Styling */
		.service-card.deleted-card {
			opacity: 0.45;
			border-color: var(--border) !important;
			box-shadow: none !important;
			transform: none !important;
		}

		.service-card.deleted-card .delete-action {
			background-color: var(--border) !important;
			color: var(--text-muted) !important;
			border-color: var(--border) !important;
			cursor: not-allowed;
			pointer-events: none;
		}

		/* Delete Button */
		.delete-action {
			display: inline-flex;
			align-items: center;
			justify-content: center;
			gap: 0.5rem;
			width: 100%;
			background-color: var(--delete-btn-bg);
			color: var(--delete-btn-text);
			border: 1px solid var(--delete-btn-border);
			border-radius: 0.75rem;
			padding: 0.65rem;
			font-family: inherit;
			font-weight: 500;
			font-size: 0.9rem;
			text-decoration: none;
			margin-top: 1rem;
			cursor: pointer;
			transition: all 0.2s;
		}

		.delete-action:hover {
			background-color: var(--delete-btn-hover-bg);
			color: #ffffff;
			border-color: var(--delete-btn-hover-border);
			box-shadow: 0 0 10px var(--accent-glow);
		}

		/* Responsive tweaks */
		@media (max-width: 640px) {
			body {
				padding: 1.5rem 1rem;
			}
			.controls-bar {
				flex-direction: column;
				align-items: stretch;
			}
			.filters {
				justify-content: center;
			}
		}
	</style>
</head>
<body>
	<div class="container" style="position: relative;">
		<div class="theme-toggle-wrapper">
			<button id="themeToggle" class="filter-btn" style="display: flex; align-items: center; gap: 0.5rem; border-radius: 2rem; padding: 0.5rem 1.25rem;">
				<span id="themeToggleIcon">🌙</span> <span id="themeToggleText">Dark Mode</span>
			</button>
		</div>
		<header>
			<h1>Digital Footprint Report</h1>
			<p class="subtitle">Identified digital services and accounts based on email history</p>
		</header>

		<!-- Statistics -->
		<section class="stats-grid">
			<div class="stat-card">
				<div class="stat-info">
					<h2>Detected Services</h2>
					<div class="stat-value">{{.Stats.Total}}</div>
				</div>
				<div class="stat-icon">🕸️</div>
			</div>
			<div class="stat-card">
				<div class="stat-info">
					<h2>High Confidence Accounts</h2>
					<div class="stat-value" style="color: #10b981;">{{.Stats.HighConf}}</div>
				</div>
				<div class="stat-icon">🛡️</div>
			</div>
			<div class="stat-card">
				<div class="stat-info">
					<h2>Subscriptions & Paid</h2>
					<div class="stat-value" style="color: #3b82f6;">{{.Stats.Subscriptions}}</div>
				</div>
				<div class="stat-icon">💳</div>
			</div>
			<div class="stat-card">
				<div class="stat-info">
					<h2>Deleted Accounts</h2>
					<div class="stat-value" id="deletedStatVal" style="color: #10b981;">0 / {{.Stats.Total}}</div>
				</div>
				<div class="stat-icon">✅</div>
			</div>
		</section>

		<!-- Filter and Search controls -->
		<section class="controls-bar">
			<div class="search-wrapper">
				<input type="text" id="searchInput" class="search-input" placeholder="Search service name or domain...">
			</div>
			<div class="filters">
				<button class="filter-btn active" data-filter="all">All</button>
				<button class="filter-btn" data-filter="high">High Confidence (7+)</button>
				<button class="filter-btn" data-filter="subscriptions">Subscriptions</button>
				<button class="filter-btn" data-filter="low">Low Confidence (&lt;4)</button>
			</div>
		</section>

		<!-- Services List -->
		<main id="servicesGrid" class="services-grid">
			{{range .Services}}
			<article class="service-card" 
					 data-name="{{.Name}}" 
					 data-domain="{{.Domain}}"
					 data-confidence="{{.Confidence}}"
					 data-receipt="{{.HasReceipt}}">
				<div class="card-header">
					<div style="display: flex; gap: 0.75rem; align-items: flex-start;">
						<label class="deleted-checkbox-wrapper" title="Mark account as deleted">
							<input type="checkbox" class="deleted-checkbox" onchange="toggleDeleted(this, '{{.Domain}}')">
							<span class="checkmark"></span>
						</label>
						<div class="service-identity">
							<h3>{{.Name}}</h3>
							<div class="service-domain">{{.Domain}}</div>
						</div>
					</div>
					<div class="confidence-badge {{if ge .Confidence 7}}conf-high{{else if ge .Confidence 4}}conf-mid{{else}}conf-low{{end}}">
						Score: {{.Confidence}}/10
					</div>
				</div>

				<div class="indicators">
					{{if .HasWelcome}}<span class="badge welcome">Welcome Email</span>{{end}}
					{{if .HasReset}}<span class="badge reset">Password Reset</span>{{end}}
					{{if .HasReceipt}}<span class="badge receipt">Payment / Invoice</span>{{end}}
					{{if and (not .HasWelcome) (and (not .HasReset) (not .HasReceipt))}}<span class="badge unclassified">General Interaction</span>{{end}}
					<span class="badge count">{{.SourcesCount}} Email{{if ne .SourcesCount 1}}s{{end}}</span>
				</div>

				<div class="verify-section">
					<button class="toggle-subjects" onclick="toggleDetails(this)">Verify Email Subjects</button>
					<ul class="subjects-list">
						{{range .SampleSubjects}}
						<li>{{.}}</li>
						{{end}}
					</ul>
				</div>

				<a href="{{.DeleteURL}}" target="_blank" rel="noopener noreferrer" class="delete-action">
					<span>Request Account Deletion</span>
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path>
						<polyline points="15 3 21 3 21 9"></polyline>
						<line x1="10" y1="14" x2="21" y2="3"></line>
					</svg>
				</a>
			</article>
			{{end}}
		</main>
	</div>

	<script>
		// Expandable details for email verification
		function toggleDetails(button) {
			button.classList.toggle('active');
			const list = button.nextElementSibling;
			list.classList.toggle('show');
		}

		// Client side filtering & searching
		const searchInput = document.getElementById('searchInput');
		const filterButtons = document.querySelectorAll('.filter-btn');
		const cards = document.querySelectorAll('.service-card');

		let currentFilter = 'all';
		let searchQuery = '';

		function applyFilters() {
			cards.forEach(card => {
				const name = card.getAttribute('data-name').toLowerCase();
				const domain = card.getAttribute('data-domain').toLowerCase();
				const confidence = parseInt(card.getAttribute('data-confidence'));
				const isReceipt = card.getAttribute('data-receipt') === 'true';

				const matchesSearch = name.includes(searchQuery) || domain.includes(searchQuery);
				let matchesFilter = false;

				if (currentFilter === 'all') {
					matchesFilter = true;
				} else if (currentFilter === 'high') {
					matchesFilter = confidence >= 7;
				} else if (currentFilter === 'subscriptions') {
					matchesFilter = isReceipt;
				} else if (currentFilter === 'low') {
					matchesFilter = confidence < 4;
				}

				if (matchesSearch && matchesFilter) {
					card.style.display = 'flex';
				} else {
					card.style.display = 'none';
				}
			});
		}

		searchInput.addEventListener('input', (e) => {
			searchQuery = e.target.value.toLowerCase().trim();
			applyFilters();
		});

		filterButtons.forEach(btn => {
			btn.addEventListener('click', (e) => {
				filterButtons.forEach(b => b.classList.remove('active'));
				btn.classList.add('active');
				currentFilter = btn.getAttribute('data-filter');
				applyFilters();
			});
		});

		// Theme Toggle Logic
		const themeToggle = document.getElementById('themeToggle');
		const themeToggleIcon = document.getElementById('themeToggleIcon');
		const themeToggleText = document.getElementById('themeToggleText');
		const body = document.body;

		let savedTheme = 'light';
		try {
			savedTheme = localStorage.getItem('theme') || 'light';
		} catch (e) {
			// ignore SecurityError in sandbox / Chrome about:blank frame
		}

		// Load preferred theme, default to light. If dark is saved, load dark.
		if (savedTheme === 'dark') {
			body.classList.add('dark-mode');
			themeToggleIcon.textContent = '☀️';
			themeToggleText.textContent = 'Light Mode';
		}

		themeToggle.addEventListener('click', () => {
			body.classList.toggle('dark-mode');
			const isDark = body.classList.contains('dark-mode');
			try {
				localStorage.setItem('theme', isDark ? 'dark' : 'light');
			} catch (e) {
				// ignore
			}
			themeToggleIcon.textContent = isDark ? '☀️' : '🌙';
			themeToggleText.textContent = isDark ? 'Light Mode' : 'Dark Mode';
		});

		// Deleted Accounts persistence logic
		const deletedStatVal = document.getElementById('deletedStatVal');

		function updateDeletedStats() {
			const totalCards = cards.length;
			let deletedCount = 0;
			cards.forEach(card => {
				const domain = card.getAttribute('data-domain');
				let isDeleted = false;
				try {
					isDeleted = localStorage.getItem('deleted-service-' + domain) === 'true';
				} catch (e) {
					// ignore
				}
				const checkbox = card.querySelector('.deleted-checkbox');
				
				if (isDeleted) {
					card.classList.add('deleted-card');
					if (checkbox) checkbox.checked = true;
					deletedCount++;
				} else {
					card.classList.remove('deleted-card');
					if (checkbox) checkbox.checked = false;
				}
			});

			if (deletedStatVal) {
				deletedStatVal.textContent = deletedCount + ' / ' + totalCards;
			}
		}

		window.toggleDeleted = function(checkbox, domain) {
			try {
				localStorage.setItem('deleted-service-' + domain, checkbox.checked ? 'true' : 'false');
			} catch (e) {
				// ignore
			}
			updateDeletedStats();
		};

		// Initial load of stats
		updateDeletedStats();
	</script>
</body>
</html>
`
