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
	<title>Digital Footprint Analyzer</title>
	<link rel="preconnect" href="https://fonts.googleapis.com">
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
	<link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;500;600;700&display=swap" rel="stylesheet">
	<style>
		:root {
			--bg-dark: #1B0C0C;
			--panel-dark: #2E1615;
			--accent: #90DD55;
			--accent-glow: rgba(144, 221, 85, 0.15);
			--text-main: #F2E4E3;
			--text-muted: #C8A679;
			--border: rgba(200, 166, 121, 0.15);
			--welcome-color: #90DD55;
			--reset-color: #C8A679;
			--receipt-color: #F2E4E3;
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
			background: linear-gradient(135deg, #F2E4E3, #90DD55, #C8A679);
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

		.stat-info h3 {
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
			opacity: 0.25;
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

		.service-identity h2 {
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

		.conf-high { background-color: rgba(144, 221, 85, 0.15); color: #90DD55; border: 1px solid rgba(144, 221, 85, 0.3); }
		.conf-mid { background-color: rgba(200, 166, 121, 0.15); color: #C8A679; border: 1px solid rgba(200, 166, 121, 0.3); }
		.conf-low { background-color: rgba(242, 228, 227, 0.1); color: #F2E4E3; border: 1px solid rgba(242, 228, 227, 0.2); }

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

		.badge.welcome { background-color: rgba(144, 221, 85, 0.1); color: var(--welcome-color); border: 1px solid rgba(144, 221, 85, 0.2); }
		.badge.reset { background-color: rgba(200, 166, 121, 0.1); color: var(--reset-color); border: 1px solid rgba(200, 166, 121, 0.2); }
		.badge.receipt { background-color: rgba(242, 228, 227, 0.1); color: var(--receipt-color); border: 1px solid rgba(242, 228, 227, 0.2); }
		.badge.unclassified { background-color: rgba(242, 228, 227, 0.05); color: var(--text-muted); border: 1px solid rgba(242, 228, 227, 0.15); }
		.badge.count { background-color: rgba(200, 166, 121, 0.1); color: var(--text-muted); border: 1px solid rgba(200, 166, 121, 0.2); }

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

		/* Delete Button */
		.delete-action {
			display: inline-flex;
			align-items: center;
			justify-content: center;
			gap: 0.5rem;
			width: 100%;
			background-color: rgba(239, 68, 68, 0.1);
			color: #ef4444;
			border: 1px solid rgba(239, 68, 68, 0.2);
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
			background-color: #ef4444;
			color: #ffffff;
			border-color: #ef4444;
			box-shadow: 0 0 10px rgba(239, 68, 68, 0.3);
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
	<div class="container">
		<header>
			<h1>Digital Footprint Report</h1>
			<p class="subtitle">Identified digital services and accounts based on email history</p>
		</header>

		<!-- Statistics -->
		<section class="stats-grid">
			<div class="stat-card">
				<div class="stat-info">
					<h3>Detected Services</h3>
					<div class="stat-value">{{.Stats.Total}}</div>
				</div>
				<div class="stat-icon">🕸️</div>
			</div>
			<div class="stat-card">
				<div class="stat-info">
					<h3>High Confidence Accounts</h3>
					<div class="stat-value" style="color: #10b981;">{{.Stats.HighConf}}</div>
				</div>
				<div class="stat-icon">🛡️</div>
			</div>
			<div class="stat-card">
				<div class="stat-info">
					<h3>Subscriptions & Paid</h3>
					<div class="stat-value" style="color: #3b82f6;">{{.Stats.Subscriptions}}</div>
				</div>
				<div class="stat-icon">💳</div>
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
					<div class="service-identity">
						<h2>{{.Name}}</h2>
						<div class="service-domain">{{.Domain}}</div>
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
	</script>
</body>
</html>
`
