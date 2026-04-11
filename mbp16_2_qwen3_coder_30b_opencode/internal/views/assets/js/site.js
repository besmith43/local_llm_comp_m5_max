let currentFolder = null;
let tvShows = [];

document.addEventListener('DOMContentLoaded', function() {
    loadFolders();
});

async function loadFolders() {
    try {
        const response = await fetch('/api/folders');
        const data = await response.json();
        
        const grid = document.getElementById('folder-grid');
        if (!data.folders || data.folders.length === 0) {
            grid.innerHTML = '<div class="col-12"><div class="alert alert-info"><h4 class="alert-heading">No videos found</h4><p>No video files were found in the source directory. Check your SOURCE environment variable.</p></div></div>';
            return;
        }
        
        grid.innerHTML = '';
        data.folders.forEach(folder => {
            const card = createFolderCard(folder);
            grid.appendChild(card);
        });
    } catch (error) {
        console.error('Error loading folders:', error);
        Swal.fire({
            icon: 'error',
            title: 'Error',
            text: 'Failed to load folders. Please try again.',
        });
    }
}

function createFolderCard(folder) {
    const card = document.createElement('div');
    card.className = 'col-6 col-md-4 col-lg-3';
    
    const videosHtml = folder.videos.map(video => `
        <div class="video-item">
            <span class="file-icon">▶</span>
            <span class="file-name">${escapeHtml(video.filename)}</span>
            <span class="file-size">(${formatFileSize(video.size)})</span>
        </div>
    `).join('\n');
    
    card.innerHTML = `
        <div class="folder-card h-100" onclick="openFolder('${escapeHtml(folder.path)}', '${escapeHtml(folder.display_name)}', ${JSON.stringify(folder.videos[0] || {})}).catch(e=>console.error(e))">
            <div class="card-body">
                <h5 class="folder-name text-truncate" title="${escapeHtml(folder.display_name)}">
                    📁 ${escapeHtml(folder.display_name)}
                </h5>
                <p class="text-muted mb-2">${folder.videos.length} video(s)</p>
                <div class="video-list mt-2">
                    ${videosHtml}
                </div>
            </div>
        </div>
    `;
    
    return card;
}

async function openFolder(folderPath, folderName, video) {
    currentFolder = {
        path: folderPath,
        name: folderName,
        video: video,
    };
    
    await loadTVShows();
    
    document.getElementById('modal-folder-title').textContent = `📁 ${escapeHtml(currentFolder.name)}`;
    const modalContent = document.getElementById('modal-content');
    modalContent.innerHTML = '';
    
    const template = document.getElementById('type-selector-template');
    const clone = template.content.cloneNode(true);
    modalContent.appendChild(clone);
    
    const modal = new bootstrap.Modal(document.getElementById('videoModal'));
    modal.show();
}

async function loadTVShows() {
    try {
        const response = await fetch('/api/tv-shows');
        const data = await response.json();
        tvShows = data.tv_shows || [];
        return tvShows;
    } catch (error) {
        console.error('Error loading TV shows:', error);
        return [];
    }
}

function showMovieForm() {
    const modalContent = document.getElementById('modal-content');
    modalContent.innerHTML = '';
    
    const template = document.getElementById('movie-form-template');
    const clone = template.content.cloneNode(true);
    modalContent.appendChild(clone);
    
    if (currentFolder.video) {
        const videoPath = currentFolder.path + '/' + currentFolder.video.filename;
        document.getElementById('video-path').value = videoPath;
    }
}

function showTVShowForm() {
    loadTVShows().then(() => {
        const modalContent = document.getElementById('modal-content');
        modalContent.innerHTML = '';
        
        const template = document.getElementById('tv-show-form-template');
        const clone = template.content.cloneNode(true);
        modalContent.appendChild(clone);
        
        const select = document.getElementById('tv-show-select');
        select.innerHTML = '<option value="">Select a TV show or leave empty to create new</option>';
        tvShows.forEach(show => {
            const option = document.createElement('option');
            option.value = escapeHtml(show);
            option.textContent = escapeHtml(show);
            select.appendChild(option);
        });
        
        if (currentFolder.video) {
            const videoPath = currentFolder.path + '/' + currentFolder.video.filename;
            document.getElementById('video-path-tv').value = videoPath;
            const episodeTitle = currentFolder.video.filename.replace(/\.[^/.]+$/, "");
            document.getElementById('episode-title').value = episodeTitle;
        }
    }).catch(error => console.error('Error loading TV shows:', error));
}

function handleShowSelectChange() {
    const select = document.getElementById('tv-show-select');
    const newShowContainer = document.getElementById('new-show-container');
    const isNewShowInput = document.getElementById('is-new-show');
    
    if (select.value === '') {
        newShowContainer.style.display = 'block';
        isNewShowInput.value = 'true';
        document.getElementById('new-show-title').focus();
    } else {
        newShowContainer.style.display = 'none';
        isNewShowInput.value = 'false';
    }
}

async function handleSubmit(event) {
    event.preventDefault();
    
    const btn = event.target.querySelector('button[type="submit"]');
    const originalText = btn.textContent;
    btn.textContent = 'Processing...';
    btn.disabled = true;
    
    try {
        let formData;
        
        if (event.target.id === 'movie-form') {
            formData = {
                video_path: document.getElementById('video-path').value,
                type: 'movie',
                movie_title: document.getElementById('movie-title').value.trim(),
                movie_year: document.getElementById('movie-year').value.trim(),
            };
        } else {
            const isShow = document.getElementById('is-new-show').value === 'true';
            formData = {
                video_path: document.getElementById('video-path-tv').value,
                type: 'tv_show',
                tv_show_title: document.getElementById('tv-show-select').value,
                is_new_show: isShow,
                new_show_title: document.getElementById('new-show-title').value.trim(),
                season_number: parseInt(document.getElementById('season-number').value),
                episode_number: parseInt(document.getElementById('episode-number').value),
                episode_title: document.getElementById('episode-title').value.trim(),
            };
        }
        
        const response = await fetch('/api/move', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(formData),
        });
        
        const data = await response.json();
        
        if (response.ok) {
            Swal.fire({
                icon: 'success',
                title: 'Success!',
                text: 'Video moved successfully.',
                timer: 1500,
                showConfirmButton: false,
            }).then(() => {
                const modal = bootstrap.Modal.getInstance(document.getElementById('videoModal'));
                if (modal) {
                    modal.hide();
                }
                loadFolders();
            });
        } else {
            Swal.fire({
                icon: 'error',
                title: 'Error',
                text: data.error || 'Failed to move video.',
            });
        }
    } catch (error) {
        console.error('Error:', error);
        Swal.fire({
            icon: 'error',
            title: 'Error',
            text: 'An error occurred. Please try again.',
        });
    } finally {
        btn.textContent = originalText;
        btn.disabled = false;
    }
}

function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i =Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
