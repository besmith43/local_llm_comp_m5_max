// Plex Importor - Main JavaScript

document.addEventListener('DOMContentLoaded', function() {
	// Initialize state
	let selectedFolder = null;
	let selectedFile = null;
	let importModal = null;

	// Initialize Bootstrap modal
	importModal = new bootstrap.Modal(document.getElementById('importModal'));

	// Load folders on page load
	loadFolders();

	// Setup event listeners
	setupEventListeners();
});

function setupEventListeners() {
	// Content type radio buttons
	document.getElementById('typeMovie').addEventListener('change', function() {
		showForm('movie');
	});

	document.getElementById('typeTVShow').addEventListener('change', function() {
		showForm('tv_show');
		loadTVShows();
	});

	// TV Show select change - toggle new show title input visibility
	document.getElementById('tvShowSelect').addEventListener('change', function() {
		const newShowInput = document.getElementById('newShowTitle');
		if (this.value === '') {
			newShowInput.classList.remove('d-none');
			newShowInput.setAttribute('required', 'true');
		} else {
			newShowInput.classList.add('d-none');
			newShowInput.removeAttribute('required');
		}
	});

	// Submit button
	document.getElementById('submitImport').addEventListener('click', submitImport);
}

function loadFolders() {
	fetch('/api/folders')
		.then(response => response.json())
		.then(data => {
			const folderGrid = document.getElementById('folderGrid');
			const loadingState = document.getElementById('loadingState');
			const emptyState = document.getElementById('emptyState');

			// Hide loading state
			loadingState.classList.remove('visible');

			if (!data.success) {
				Swal.fire({
					icon: 'error',
					title: 'Error',
					text: data.error || 'Failed to load folders',
				});
				return;
			}

			if (data.folders.length === 0) {
				emptyState.classList.add('visible');
				return;
			}

			// Render folders
			folderGrid.innerHTML = '';
			data.folders.forEach(folder => {
				const card = createFolderCard(folder);
				folderGrid.appendChild(card);
			});
		})
		.catch(error => {
			console.error('Error loading folders:', error);
			document.getElementById('loadingState').classList.remove('visible');
			Swal.fire({
				icon: 'error',
				title: 'Error',
				text: 'Failed to load folders. Please check the console for details.',
			});
		});
}

function createFolderCard(folder) {
	const card = document.createElement('div');
	card.className = 'folder-card';
	card.dataset.folderPath = folder.path;

	let videoFilesHtml = '';
	folder.video_files.forEach(file => {
		const fileSize = formatFileSize(file.size);
		videoFilesHtml += `
			<div class="video-file-item" data-filename="${file.name}" data-filepath="${file.path}">
				<i class="fa-solid fa-film video-file-icon"></i>
				<span class="video-file-name" title="${file.name}">${file.name}</span>
				<span class="video-file-size">${fileSize}</span>
			</div>
		`;
	});

	const totalSize = formatFileSize(folder.total_size);

	card.innerHTML = `
		<div class="card-body">
			<i class="fa-solid fa-folder folder-icon"></i>
			<div class="folder-name" title="${folder.name}">${folder.name}</div>
			<div class="folder-meta">
				${folder.file_count} video file${folder.file_count !== 1 ? 's' : ''} | ${totalSize}
			</div>
			<div class="video-files-list">
				${videoFilesHtml}
			</div>
		</div>
	`;

	// Click on video file item to select that specific file
	card.querySelectorAll('.video-file-item').forEach(item => {
		item.style.cursor = 'pointer';
		item.addEventListener('click', function(e) {
			e.stopPropagation();
			selectFile(folder, this.dataset.filename, this.dataset.filepath);
		});
	});

	// Click on card to select first video file
	card.addEventListener('click', function() {
		if (folder.video_files.length > 0) {
			selectFile(folder, folder.video_files[0].name, folder.video_files[0].path);
		}
	});

	return card;
}

function selectFile(folder, filename, filepath) {
	selectedFolder = folder;
	selectedFile = { name: filename, path: filepath };

	// Set hidden form fields
	document.getElementById('sourceType').value = '';
	document.getElementById('sourcePath').value = filepath;

	// Update modal info
	document.getElementById('selectedFileName').textContent = filename;
	document.getElementById('selectedFolder').textContent = folder.path;

	// Reset form
	resetForm();

	// Show modal
	importModal.show();
}

function resetForm() {
	// Reset content type selection to movie
	document.getElementById('typeMovie').checked = true;
	showForm('movie');

	// Reset all inputs
	document.getElementById('movieTitle').value = '';
	document.getElementById('movieYear').value = '';
	document.getElementById('tvShowSelect').value = '';
	document.getElementById('newShowTitle').value = '';
	document.getElementById('seasonNumber').value = '';
	document.getElementById('episodeNumber').value = '';

	// Hide new show title input
	document.getElementById('newShowTitle').classList.add('d-none');
}

function showForm(type) {
	const movieForm = document.getElementById('movieForm');
	const tvShowForm = document.getElementById('tvShowForm');

	if (type === 'movie') {
		movieForm.classList.remove('d-none');
		tvShowForm.classList.add('d-none');
	} else {
		movieForm.classList.add('d-none');
		tvShowForm.classList.remove('d-none');
	}
}

function loadTVShows() {
	fetch('/api/tv-shows')
		.then(response => response.json())
		.then(data => {
			const select = document.getElementById('tvShowSelect');

			if (!data.success) {
				console.error('Error loading TV shows:', data.error);
				return;
			}

			// Clear existing options except the first one
			select.innerHTML = '<option value="">-- Select or create new --</option>';

			// Add TV show options
			data.shows.forEach(show => {
				const option = document.createElement('option');
				option.value = show;
				option.textContent = show;
				select.appendChild(option);
			});

			// If no shows exist, show a message
			if (data.shows.length === 0) {
				const option = document.createElement('option');
				option.value = '__empty__';
				option.textContent = '(No existing shows - enter new show title below)';
				select.appendChild(option);
			}
		})
		.catch(error => {
			console.error('Error loading TV shows:', error);
		});
}

function submitImport() {
	const sourceType = document.querySelector('input[name="contentType"]:checked').value;

	// Validate based on content type
	if (sourceType === 'movie') {
		if (!validateMovieForm()) {
			return;
		}

		const payload = {
			source_type: 'movie',
			source_path: selectedFile.path,
			movie_title: document.getElementById('movieTitle').value.trim(),
			movie_year: parseInt(document.getElementById('movieYear').value),
		};

		sendImportRequest(payload);
	} else {
		if (!validateTVShowForm()) {
			return;
		}

		const tvShowTitle = document.getElementById('tvShowSelect').value;
		const newShowTitle = document.getElementById('newShowTitle').value.trim();

		const payload = {
			source_type: 'tv_show',
			source_path: selectedFile.path,
			tv_show_title: tvShowTitle || '',
			new_show_title: newShowTitle,
			season_number: parseInt(document.getElementById('seasonNumber').value),
			episode_number: parseInt(document.getElementById('episodeNumber').value),
		};

		sendImportRequest(payload);
	}
}

function validateMovieForm() {
	const title = document.getElementById('movieTitle').value.trim();
	const year = document.getElementById('movieYear').value;

	if (!title) {
		Swal.fire({
			icon: 'warning',
			title: 'Missing Field',
			text: 'Please enter a movie title.',
		});
		document.getElementById('movieTitle').focus();
		return false;
	}

	if (!year || year < 1888 || year > 2030) {
		Swal.fire({
			icon: 'warning',
			title: 'Invalid Year',
			text: 'Please enter a valid year (1888-2030).',
		});
		document.getElementById('movieYear').focus();
		return false;
	}

	return true;
}

function validateTVShowForm() {
	const tvShowTitle = document.getElementById('tvShowSelect').value;
	const newShowTitle = document.getElementById('newShowTitle').value.trim();
	const season = document.getElementById('seasonNumber').value;
	const episode = document.getElementById('episodeNumber').value;

	// Check if a show is selected or new title provided
	if (!tvShowTitle && !newShowTitle) {
		Swal.fire({
			icon: 'warning',
			title: 'Missing Field',
			text: 'Please select an existing TV show or enter a new show title.',
		});
		if (document.getElementById('tvShowSelect').options.length > 2) {
			document.getElementById('tvShowSelect').focus();
		} else {
			document.getElementById('newShowTitle').focus();
		}
		return false;
	}

	if (!season || season < 1) {
		Swal.fire({
			icon: 'warning',
			title: 'Invalid Season',
			text: 'Please enter a valid season number (1 or greater).',
		});
		document.getElementById('seasonNumber').focus();
		return false;
	}

	if (!episode || episode < 1) {
		Swal.fire({
			icon: 'warning',
			title: 'Invalid Episode',
			text: 'Please enter a valid episode number (1 or greater).',
		});
		document.getElementById('episodeNumber').focus();
		return false;
	}

	return true;
}

function sendImportRequest(payload) {
	fetch('/api/move', {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
		},
		body: JSON.stringify(payload),
	})
		.then(response => response.json())
		.then(data => {
			if (data.success) {
				Swal.fire({
					icon: 'success',
					title: 'Success!',
					text: data.message || 'File imported successfully.',
					timer: 2000,
					showConfirmButton: false,
				}).then(() => {
					importModal.hide();
					loadFolders(); // Refresh the folder list
				});
			} else {
				Swal.fire({
					icon: 'error',
					title: 'Error',
					text: data.error || 'Failed to import file.',
				});
			}
		})
		.catch(error => {
			console.error('Error importing file:', error);
			Swal.fire({
				icon: 'error',
				title: 'Error',
				text: 'Failed to import file. Please check the console for details.',
			});
		});
}

function formatFileSize(bytes) {
	if (bytes === 0) return '0 B';

	const k = 1024;
	const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
	const i = Math.floor(Math.log(bytes) / Math.log(k));

	return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}
