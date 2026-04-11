document.addEventListener('DOMContentLoaded', () => {
  const scanBtn = document.getElementById('scanBtn');
  const foldersContainer = document.getElementById('foldersContainer');
  const modalOverlay = document.getElementById('modalOverlay');
  const modalContent = document.getElementById('modalContent');
  const closeModal = () => { modalOverlay.classList.add('hidden'); };
  const importForm = document.getElementById('importForm');
  const errorMsg = document.getElementById('errorMsg');
  const movieFields = document.getElementById('movieFields');
  const tvFields = document.getElementById('tvFields');

  // Toggle fields based on type
  importForm.type.addEventListener('change', () => {
    if (importForm.type.value === 'movie') {
      movieFields.classList.remove('hidden');
      tvFields.classList.add('hidden');
    } else {
      movieFields.classList.add('hidden');
      tvFields.classList.remove('hidden');
    }
  });

  // Scan source directory
  scanBtn.addEventListener('click', async () => {
    try {
      const resp = await fetch('/scan', { method: 'POST' });
      const folders = await resp.json();
      foldersContainer.innerHTML = '';
      folders.forEach(folder => {
        const div = document.createElement('div');
        div className = 'folder';
        div.textContent = folder.folder_name;
        div.onclick = () => openModal(folder);
        foldersContainer.appendChild(div);
      });
    } catch (e) {
      console.error('Scan failed', e);
    }
  });

  // Open modal and pre-fill data
  const openModal = async (folder) => {
    modalOverlay.classList.remove('hidden');
    // Populate series dropdown for TV shows
    if (importForm.type.value === 'tv_show') {
      const seriesResp = await fetch('/series');
      const seriesList = await seriesResp.json();
      const seriesSelect = document.getElementById('seriesSelect');
      seriesSelect.innerHTML = '<option value="">-- Select Series --</option>';
      seriesList.forEach(s => {
        const opt = document.createElement('option');
        opt.value = s;
        opt.textContent = s;
        seriesSelect.appendChild(opt);
      });
    }
  };

  // Handle form submission
  importForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    errorMsg.classList.add('hidden');
    errorMsg.textContent = '';

    const formData = new FormData(importForm);
    const payload = {
      type: importForm.type.value,
      title: importForm.title?.value?.trim() || '',
      year: importForm.year?.value ? parseInt(importForm.year.value) : 0,
      series: importForm.series?.value?.trim() || '',
      season: importForm.season?.value ? parseInt(importForm.season.value) : 0,
      episode: importForm.episode?.value ? parseInt(importForm.episode.value) : 0,
      source_path: '', // will be filled later based on selected folder
      ext: '.mp4' // placeholder; will be adjusted if needed
    };

    // Determine selected folder name to construct source_path
    // The folder name is stored as data-folder attribute on the clicked folder div
    const selectedFolder = document.querySelector('.folder.active');
    if (!selectedFolder) {
      errorMsg.textContent = 'Please select a folder first';
      errorMsg.classList.remove('hidden');
      return;
    }
    payload.source_path = selectedFolder.dataset.folder;

    // Adjust extension based on actual file extension (could be determined from file listing)
    // For simplicity, assume .mp4; can be changed later.

    // Special handling for TV show: ensure season/episode are numbers
    if (payload.type === 'tv_show') {
      if (!payload.series || !payload.season || !payload.episode) {
        errorMsg.textContent = 'Series, season, and episode are required';
        errorMsg.classList.remove('hidden');
        return;
      }
    }

    // Send to /move
    try {
      const resp = await fetch('/move', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      const result = await resp.json();
      if (result.status === 'moved') {
        alert('File moved successfully');
        closeModal();
      } else {
        errorMsg.textContent = result.error || 'Move failed';
        errorMsg.classList.remove('hidden');
      }
    } catch (err) {
      console.error('Move error', err);
      errorMsg.textContent = 'Request failed';
      errorMsg.classList.remove('hidden');
    }
  });
});
