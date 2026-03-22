// Global state
let currentSourceDir = '';
let currentFilename = '';
let currentType = '';

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    loadDirectories();
});

// Load directories from API
function loadDirectories() {
    fetch('/api/directories')
        .then(response => response.json())
        .then(data => {
            const grid = document.getElementById('directoryGrid');
            if (data.directories.length === 0) {
                grid.innerHTML = '<div class="col-12 text-center"><p class="text-muted">No video files found in source directory</p></div>';
                return;
            }

            grid.innerHTML = data.directories.map(dir => `
                <div class="col">
                    <div class="card h-100 cursor-pointer" onclick="selectSourceDir('${dir}')">
                        <div class="card-body text-center">
                            <i class="fa-solid fa-folder fa-3x mb-2 text-primary"></i>
                            <h6 class="card-title">${escapeHtml(dir)}</h6>
                        </div>
                    </div>
                </div>
            `).join('');
        })
        .catch(error => {
            console.error('Error loading directories:', error);
            Swal.fire({
                icon: 'error',
                title: 'Error',
                text: 'Failed to load directories'
            });
        });
}

// Select a source directory
function selectSourceDir(dir) {
    currentSourceDir = dir;
    // Get the first video file in this directory
    getFirstVideoFile(dir, (filename) => {
        currentFilename = filename;
        document.getElementById('typeModal').dataset.sourceDir = dir;
        const modal = new bootstrap.Modal(document.getElementById('typeModal'));
        modal.show();
    });
}

// Get the first video file from a directory
function getFirstVideoFile(dir, callback) {
    fetch(`/api/files?dir=${encodeURIComponent(dir)}`)
        .then(response => response.json())
        .then(data => {
            if (data.files && data.files.length > 0) {
                callback(data.files[0]);
            } else {
                Swal.fire({
                    icon: 'error',
                    title: 'No Files',
                    text: 'No video files found in this directory'
                });
            }
        })
        .catch(error => {
            console.error('Error loading files:', error);
            Swal.fire({
                icon: 'error',
                title: 'Error',
                text: 'Failed to load files from directory'
            });
        });
}

// Select content type (movie or TV show)
function selectType(type) {
    currentType = type;
    const sourceDir = document.getElementById('typeModal').dataset.sourceDir;

    if (!sourceDir) {
        Swal.fire({
            icon: 'error',
            title: 'Error',
            text: 'No directory selected'
        });
        return;
    }

    getFirstVideoFile(sourceDir, (filename) => {
        currentFilename = filename;

        if (type === 'movie') {
            document.getElementById('movieSourceDir').value = sourceDir;
            document.getElementById('movieFilename').value = filename;
            const modal = new bootstrap.Modal(document.getElementById('movieModal'));

            // Hide the type selection modal first
            const typeModalEl = document.getElementById('typeModal');
            const typeModal = bootstrap.Modal.getInstance(typeModalEl);
            if (typeModal) {
                typeModal.hide();
            }

            setTimeout(() => {
                modal.show();
            }, 200);
        } else {
            // Get TV shows for the dropdown
            loadTVShows((shows) => {
                const tvSelect = document.getElementById('seriesTitle');
                tvSelect.innerHTML = '<option value="">-- Select a series --</option>';
                shows.forEach(show => {
                    const option = document.createElement('option');
                    option.value = show;
                    option.textContent = show;
                    tvSelect.appendChild(option);
                });

                // Add "Create New Series" option
                const newOption = document.createElement('option');
                newOption.value = '-- Create New Series --';
                newOption.textContent = '+ Create New Series...';
                tvSelect.appendChild(newOption);

                document.getElementById('tvSourceDir').value = sourceDir;
                document.getElementById('tvFilename').value = filename;

                const modal = new bootstrap.Modal(document.getElementById('tvShowModal'));

                // Hide the type selection modal first
                const typeModalEl = document.getElementById('typeModal');
                const typeModal = bootstrap.Modal.getInstance(typeModalEl);
                if (typeModal) {
                    typeModal.hide();
                }

                setTimeout(() => {
                    modal.show();
                }, 200);
            });
        }
    });
}

// Load TV shows for dropdown
function loadTVShows(callback) {
    fetch('/api/tv-shows')
        .then(response => response.json())
        .then(data => {
            callback(data.shows);
        })
        .catch(error => {
            console.error('Error loading TV shows:', error);
            callback([]);
        });
}

// Handle series selection change (toggle new series option)
function onSeriesSelect() {
    const select = document.getElementById('seriesTitle');
    const newSeriesGroup = document.getElementById('newSeriesGroup');

    if (select.value === '' || select.value === '-- Create New Series --') {
        newSeriesGroup.classList.remove('d-none');
    } else {
        newSeriesGroup.classList.add('d-none');
    }
}

// Submit movie form
function submitMovie() {
    const title = document.getElementById('movieTitle').value.trim();
    const year = document.getElementById('movieYear').value.trim();
    const sourceDir = document.getElementById('movieSourceDir').value;
    let filename = document.getElementById('movieFilename').value;

    // Get actual extension from the stored filename
    const extMatch = filename.match(/\.[^/.]+$/);
    const extension = extMatch ? extMatch[0].substring(1) : 'mp4';

    if (!title) {
        Swal.fire({
            icon: 'warning',
            title: 'Missing Title',
            text: 'Please enter a movie title'
        });
        return;
    }

    if (!year || year.length !== 4 || year < 1900 || year > new Date().getFullYear() + 5) {
        Swal.fire({
            icon: 'warning',
            title: 'Invalid Year',
            text: 'Please enter a valid 4-digit year'
        });
        return;
    }

    const formData = {
        sourceDir: sourceDir,
        filename: filename,
        title: title,
        year: year,
        extension: extension
    };

    fetch('/api/move-movie', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(formData)
    })
    .then(response => response.json())
    .then(data => {
        if (data.message) {
            Swal.fire({
                icon: 'success',
                title: 'Success!',
                text: data.message,
                timer: 2000,
                showConfirmButton: false
            });

            // Refresh directory list
            loadDirectories();

            // Close modal
            const modal = bootstrap.Modal.getInstance(document.getElementById('movieModal'));
            if (modal) {
                modal.hide();
            }
        } else {
            throw new Error(data.error || 'Unknown error');
        }
    })
    .catch(error => {
        console.error('Error moving movie:', error);
        Swal.fire({
            icon: 'error',
            title: 'Error',
            text: error.message || 'Failed to move movie'
        });
    });
}

// Submit TV show form
function submitTVShow() {
    const seriesTitle = document.getElementById('seriesTitle').value;
    const newSeriesName = document.getElementById('newSeriesName').value.trim();
    const season = parseInt(document.getElementById('tvSeason').value);
    const episode = parseInt(document.getElementById('tvEpisode').value);
    const sourceDir = document.getElementById('tvSourceDir').value;
    let filename = document.getElementById('tvFilename').value;

    // Get actual extension from the stored filename
    const extMatch = filename.match(/\.[^/.]+$/);
    const extension = extMatch ? extMatch[0].substring(1) : 'mp4';

    // Determine if creating new series
    const select = document.getElementById('seriesTitle');
    const isNewSeries = select.value === '' || select.value === '-- Create New Series --' || newSeriesName !== '';

    if (isNewSeries && !newSeriesName) {
        Swal.fire({
            icon: 'warning',
            title: 'Missing Series Name',
            text: 'Please enter a name for the new series'
        });
        return;
    }

    if (!isNewSeries && !seriesTitle) {
        Swal.fire({
            icon: 'warning',
            title: 'Missing Series',
            text: 'Please select or create a series'
        });
        return;
    }

    if (!season || season < 1) {
        Swal.fire({
            icon: 'warning',
            title: 'Invalid Season',
            text: 'Please enter a valid season number'
        });
        return;
    }

    if (!episode || episode < 1) {
        Swal.fire({
            icon: 'warning',
            title: 'Invalid Episode',
            text: 'Please enter a valid episode number'
        });
        return;
    }

    const formData = {
        sourceDir: sourceDir,
        filename: filename,
        seriesTitle: isNewSeries ? newSeriesName : seriesTitle,
        season: season,
        episode: episode,
        newSeries: isNewSeries,
        extension: extension
    };

    fetch('/api/move-tvshow', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(formData)
    })
    .then(response => response.json())
    .then(data => {
        if (data.message) {
            Swal.fire({
                icon: 'success',
                title: 'Success!',
                text: data.message,
                timer: 2000,
                showConfirmButton: false
            });

            // Refresh directory list and TV shows
            loadDirectories();
            loadTVShows(() => {});

            // Close modal
            const modal = bootstrap.Modal.getInstance(document.getElementById('tvShowModal'));
            if (modal) {
                modal.hide();
            }
        } else {
            throw new Error(data.error || 'Unknown error');
        }
    })
    .catch(error => {
        console.error('Error moving TV show:', error);
        Swal.fire({
            icon: 'error',
            title: 'Error',
            text: error.message || 'Failed to move episode'
        });
    });
}

// Escape HTML to prevent XSS
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
