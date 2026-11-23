var catChart = null;

function loadByCategory(from, to) {
    var url = '/api/stats/by-category';
    var params = [];
    if (from) { params.push('from=' + encodeURIComponent(from)); }
    if (to)   { params.push('to='   + encodeURIComponent(to)); }
    if (params.length) { url += '?' + params.join('&'); }

    fetch(url)
        .then(function (r) {
            if (!r.ok) { throw new Error('HTTP ' + r.status); }
            return r.json();
        })
        .then(function (data) {
            var labels = data.map(function (x) { return x.name; });
            var amounts = data.map(function (x) { return x.amount; });

            if (catChart) {
                // Обновление существующей диаграммы
                catChart.data.labels = labels;
                catChart.data.datasets[0].data = amounts;
                catChart.update();
                return;
            }

            catChart = new Chart(document.getElementById('catChart'), {
                type: 'pie',
                data: {
                    labels: labels,
                    datasets: [{ label: 'Сумма', data: amounts }]
                },
                options: {
                    responsive: true,
                    plugins: { legend: { position: 'bottom' } }
                }
            });
        })
        .catch(function (err) { console.error('by-category:', err); });
}


var monthChart = null;

function loadByMonth(year) {
    fetch('/api/stats/by-month?year=' + encodeURIComponent(year))
        .then(function (r) {
            if (!r.ok) { throw new Error('HTTP ' + r.status); }
            return r.json();
        })
        .then(function (data) {
            var labels = data.map(function (x) { return x.month; });
            var amounts = data.map(function (x) { return x.amount; });

            if (monthChart) {
                monthChart.data.labels = labels;
                monthChart.data.datasets[0].data = amounts;
                monthChart.update();
                return;
            }

            monthChart = new Chart(document.getElementById('monthChart'), {
                type: 'bar',
                data: {
                    labels: labels,
                    datasets: [{ label: 'Сумма за месяц', data: amounts }]
                },
                options: {
                    responsive: true,
                    scales: { y: { beginAtZero: true } }
                }
            });
        })
        .catch(function (err) { console.error('by-month:', err); });
}

// --- Инициализация при загрузке страницы ---
document.addEventListener('DOMContentLoaded', function () {
    loadByCategory('', '');

    var yearSelect = document.getElementById('year');
    if (yearSelect && yearSelect.value) {
        loadByMonth(yearSelect.value);
    }

    if (yearSelect) {
        yearSelect.addEventListener('change', function () {
            if (yearSelect.value) { loadByMonth(yearSelect.value); }
        });
    }

    var applyBtn = document.getElementById('applyPeriod');
    if (applyBtn) {
        applyBtn.addEventListener('click', function () {
            var from = document.getElementById('from').value;
            var to = document.getElementById('to').value;
            loadByCategory(from, to);
        });
    }
});