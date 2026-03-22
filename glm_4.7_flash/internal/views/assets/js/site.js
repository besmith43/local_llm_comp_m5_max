// Global variables
let selectedFolder = null;
let selectedFolderPath = null;
let selectedVideoFile = null;

// Load configuration on page load
document.addEventListener('DOMContentLoaded', function() {
    loadConfiguration();
    loadFolders();
    setupEventListeners();
});

// Load configuration from server
async function loadConfiguration() {
    try {
        const response = await fetch('/api/config');
        const config = await response.json();
        document.getElementById('sourcePath').textContent = config.sourcePath;
        document.getElementById('destinationPath').textContent = config.destinationPath;
    } catch (error) {
        console.error('Error loading configuration:', error);
    }
}

// Load folders from server
async function loadFolders() {
    const folderList = document.getElementById('folderList');
    folderList.innerHTML = `
        <div class="col-12 text-center py-5">
            <div class="spinner-border text-primary" role="status">
                <span class="visually-hidden">Loading...</span>
            </div>
            <p class="mt-3">Loading folders...</p>
        </div>
    `;

    try {
        const response = await fetch('/api/folders');
        const folders = await response.json();

        if (folders.length === 0) {
            folderList.innerHTML = `
                <div class="col-12 text-center py-5">
                    <p class="text-muted">No folders with video files found.</p>
                </div>
            `;
            return;
        }

        folderList.innerHTML = '';
        folders.forEach(folder => {
            const card = templ.html(FolderCard(folder.name, folder.path, folder.videoFile));
            folderList.innerHTML += card;
        });
    } catch (error) {
        console.error('Error loading folders:', error);
        folderList.innerHTML = `
            <div class="col-12 text-center py-5">
                <p class="text-danger">Error loading folders. Please try again.</p>
            </div>
        `;
    }
}

// Setup event listeners
function setupEventListeners() {
    // Content type selection
    document.querySelectorAll('input[name="contentType"]').forEach(radio => {
        radio.addEventListener('change', handleContentTypeChange);
    });

    // New series checkbox
    document.getElementById('newSeries').addEventListener('change', function() {
        document.getElementById('newSeriesInput').style.display = this.checked ? 'block' : 'none';
    });

    // Submit button
    document.getElementById('submitMetadata').addEventListener('click', handleSubmit);
}

// Handle content type change
function handleContentTypeChange(e) {
    const contentType = e.target.value;
    const movieForm = document.getElementById('movieForm');
    const tvForm = document.getElementById('tvForm');

    if (contentType === 'movie') {
        movieForm.style.display = 'block';
        tvForm.style.display = 'none';
    } else {
        movieForm.style.display = 'none';
        tvForm.style.display = 'block';
    }
}

// Open metadata modal
function openMetadataModal(folderName, folderPath, videoFile) {
    selectedFolder = folderName;
    selectedFolderPath = folderPath;
    selectedVideoFile = videoFile;

    document.getElementById('selectedFolder').value = folderPath;
    document.getElementById('metadataModalLabel').textContent = `Add Metadata: ${folderName}`;
    document.getElementById('movieTitle').value = '';
    document.getElementById('movieYear').value = '';
    document.getElementById('tvSeries').value = '';
    document.getElementById('newSeriesName').value = '';
    document.getElementById('tvSeason').value = '';
    document.getElementById('tvEpisode').value = '';
    document.getElementById('newSeries').checked = false;
    document.getElementById('newSeriesInput').style.display = 'none';
    document.getElementById('movieForm').style.display = 'block';
    document.getElementById('tvForm').style.display = 'none';

    // Reset content type selection
    document.querySelectorAll('input[name="contentType"]').forEach(radio => {
        radio.checked = false;
    });

    const modal = new bootstrap.Modal(document.getElementById('metadataModal'));
    modal.show();
}

// Handle form submission
async function handleSubmit() {
    const contentType = document.querySelector('input[name="contentType"]:checked');
    if (!contentType) {
        Swal.fire({
            icon: 'warning',
            title: 'Please select content type',
            text: 'Choose either Movie or TV Show',
            confirmButtonText: 'OK'
        });
        return;
    }

    const formData = {
        contentType: contentType.value,
        folderPath: selectedFolderPath
    };

    // Validate based on content type
    if (contentType.value === 'movie') {
        const title = document.getElementById('movieTitle').value.trim();
        const year = document.getElementById('movieYear').value.trim();

        if (!title || !year) {
            Swal.fire({
                icon: 'warning',
                title: 'Missing required fields',
                text: 'Please fill in both title and year',
                confirmButtonText: 'OK'
            });
            return;
        }

        formData.title = title;
        formData.year = year;
    } else {
        const seriesSelect = document.getElementById('tvSeries');
        const newSeries = document.getElementById('newSeries').checked;
        const season = document.getElementById('tvSeason').value.trim();
        const episode = document.getElementById('tvEpisode').value.trim();

        if (!season || !episode) {
            Swal.fire({
                icon: 'warning',
                title: 'Missing required fields',
                text: 'Please fill in season and episode',
                confirmButtonText: 'OK'
            });
            return;
        }

        if (newSeries) {
            const newSeriesName = document.getElementById('newSeriesName').value.trim();
            if (!newSeriesName) {
                Swal.fire({
                    icon: 'warning',
                    title: 'Missing series name',
                    text: 'Please enter a series name',
                    confirmButtonText: 'OK'
                });
                return;
            }
            formData.seriesName = newSeriesName;
        } else {
            if (!seriesSelect.value) {
                Swal.fire({
                    icon: 'warning',
                    title: 'Missing series',
                    text: 'Please select a series',
                    confirmButtonText: 'OK'
                });
                return;
            }
            formData.seriesName = seriesSelect.value;
        }

        formData.season = season;
        formData.episode = episode;
    }

    // Show loading state
    const submitBtn = document.getElementById('submitMetadata');
    submitBtn.disabled = true;
    submitBtn.textContent = 'Processing...';

    try {
        const response = await fetch('/api/metadata', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(formData)
        });

        const result = await response.json();

        if (response.ok && result.success) {
            Swal.fire({
                icon: 'success',
                title: 'Success!',
                text: result.message,
                confirmButtonText: 'OK'
            }).then(() => {
                // Close modal
                const modal = bootstrap.Modal.getInstance(document.getElementById('metadataModal'));
                modal.hide();

                // Reload folders
                loadFolders();
            });
        } else {
            Swal.fire({
                icon: 'error',
                title: 'Error',
                text: result.message || 'Failed to process metadata',
                confirmButtonText: 'OK'
            });
        }
    } catch (error) {
        console.error('Error submitting metadata:', error);
        Swal.fire({
            icon: 'error',
            title: 'Error',
            text: 'An error occurred while processing your request',
            confirmButtonText: 'OK'
        });
    } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = 'Submit';
    }
}

// Load TV series list
async function loadTVSeries() {
    try {
        const response = await fetch('/api/tv-series');
        const seriesList = await response.json();

        const seriesSelect = document.getElementById('tvSeries');
        seriesSelect.innerHTML = '<option value="">Select a series...</option>';

        seriesList.forEach(series => {
            const option = document.createElement('option');
            option.value = series;
            option.textContent = series;
            seriesSelect.appendChild(option);
        });
    } catch (error) {
        console.error('Error loading TV series:', error);
    }
}

// Load seasons for a series
async function loadSeasons(seriesName) {
    try {
        const response = await fetch(`/api/seasons/${encodeURIComponent(seriesName)}`);
        const seasons = await response.json();

        const seasonInput = document.getElementById('tvSeason');
        seasonInput.value = '';
        seasonInput.max = seasons.length.toString();

        if (seasons.length > 0) {
            seasonInput.placeholder = `1-${seasons.length}`;
        } else {
            seasonInput.placeholder = '1';
        }
    } catch (error) {
        console.error('Error loading seasons:', error);
    }
}

// Load episodes for a season
async function loadEpisodes(seriesName, season) {
    try {
        const response = await fetch(`/api/episodes/${encodeURIComponent(seriesName)}/${encodeURIComponent(season)}`);
        const episodes = await response.json();

        const episodeInput = document.getElementById('tvEpisode');
        episodeInput.value = '';
        episodeInput.max = episodes.length.toString();

        if (episodes.length > 0) {
            episodeInput.placeholder = `1-${episodes.length}`;
        } else {
            episodeInput.placeholder = '1';
        }
    } catch (error) {
        console.error('Error loading episodes:', error);
    }
}