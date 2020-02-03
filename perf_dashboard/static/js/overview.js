var brandPrimary = 'rgba(52, 168, 85,1)';

new Chart(document.getElementById('pieChart'), {
    type: 'doughnut',
    data: {
        labels: [
            "Success",
            "Failure",
            "Others"
        ],
    datasets: [
        {
            data: [450, 50, 80],
            borderWidth: [1, 1, 1],
            backgroundColor: [
                brandPrimary,
                "rgba(236, 66, 53,1)",
                "rgba(259, 188, 5,1)"
            ],
            hoverBackgroundColor: [
                brandPrimary,
                "rgba(236, 66, 53,1)",
                "rgba(259, 188, 5,1)"
            ]
        }]
    }
});