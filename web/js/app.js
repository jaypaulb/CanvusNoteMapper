// --- Tab Switching ---
const tabBtns = document.querySelectorAll('.tab-btn');
const tabSections = document.querySelectorAll('.tab-section');
tabBtns.forEach(btn => {
    btn.addEventListener('click', () => {
        tabBtns.forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        tabSections.forEach(sec => sec.style.display = 'none');
        document.getElementById(`${btn.dataset.tab}-section`).style.display = '';
    });
});
document.getElementById('tab-config').classList.add('active');

// --- Dark Mode Toggle ---
const body = document.body;
const darkmodeToggle = document.getElementById('darkmode-toggle');
const darkmodeIcon = document.getElementById('darkmode-icon');
function setDarkMode(isDark) {
    if (isDark) {
        body.classList.remove('light');
        body.classList.add('dark');
        darkmodeIcon.textContent = '🌙';
    } else {
        body.classList.remove('dark');
        body.classList.add('light');
        darkmodeIcon.textContent = '☀️';
    }
    localStorage.setItem('darkmode', isDark ? '1' : '0');
}
darkmodeToggle.addEventListener('click', () => {
    setDarkMode(!body.classList.contains('dark'));
});
// Default to dark mode
setDarkMode(localStorage.getItem('darkmode') !== '0');

// --- Persist and Restore Credentials ---
window.addEventListener('DOMContentLoaded', () => {
    const mcsServer = localStorage.getItem('mcsServer');
    const apiKey = localStorage.getItem('apiKey');
    if (mcsServer) document.getElementById('mcs-server').value = mcsServer;
    if (apiKey) document.getElementById('api-key').value = apiKey;
});

// --- Credentials ---
const credentialsForm = document.getElementById('credentials-form');
const credentialsStatus = document.getElementById('credentials-status');
credentialsForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const mcsServer = document.getElementById('mcs-server').value;
    const apiKey = document.getElementById('api-key').value;
    localStorage.setItem('mcsServer', mcsServer);
    localStorage.setItem('apiKey', apiKey);
    const res = await fetch('/api/set-credentials', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ mcsServer, apiKey })
    });
    const data = await res.json();
    credentialsStatus.textContent = data.status === 'ok' ? 'Credentials saved.' : (data.error || 'Error');
});

// --- Fetch Zones ---
const fetchZonesBtn = document.getElementById('fetch-zones');
const canvasSelect = document.getElementById('canvas-select');
const zoneSelect = document.getElementById('zone-select');
const zoneStatus = document.getElementById('zone-status');
fetchZonesBtn.addEventListener('click', async () => {
    const res = await fetch('/api/get-zones');
    if (!res.ok) {
        const data = await res.json();
        zoneStatus.textContent = data.error || 'Error fetching zones.';
        return;
    }
    const data = await res.json();
    canvasSelect.innerHTML = '';
    data.canvases.forEach(c => {
        const opt = document.createElement('option');
        opt.value = c; opt.textContent = c;
        canvasSelect.appendChild(opt);
    });
    zoneSelect.innerHTML = '';
    data.zones.forEach(z => {
        const opt = document.createElement('option');
        opt.value = z; opt.textContent = z;
        zoneSelect.appendChild(opt);
    });
    zoneStatus.textContent = 'Zones loaded.';
});

// --- Image Capture/Upload & Auto-Scan ---
const imageInput = document.getElementById('image-input');
const imageStatus = document.getElementById('image-status');
const preview = document.getElementById('preview');
let uploadedImage = null;
let lastScanData = null;
let selectedNotes = [];
const imageInputLabel = document.getElementById('image-input-label');
// imageInputLabel.addEventListener('click', () => imageInput.click());
imageInput.addEventListener('change', async (e) => {
    const file = e.target.files[0];
    if (file) {
        const reader = new FileReader();
        reader.onload = function(evt) {
            preview.src = evt.target.result;
            preview.style.display = 'block';
        };
        reader.readAsDataURL(file);
        uploadedImage = file;
        // Auto-upload
        const formData = new FormData();
        formData.append('image', uploadedImage);
        const res = await fetch('/api/upload-image', {
            method: 'POST',
            body: formData
        });
        const data = await res.json();
        imageStatus.textContent = data.status || data.error || 'Upload error.';
        if (data.status === 'uploaded') {
            // Auto-trigger scan
            autoScanNotes();
        }
    }
    // Reset file input value to allow re-selecting the same file and prevent double dialog
    imageInput.value = "";
});

async function autoScanNotes() {
    // For the stub, just send empty imageData and mock dimensions
    const res = await fetch('/api/scan-notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            imageData: [],
            imageDimensions: [640, 480],
            zoneDimensions: [640, 480]
        })
    });
    const data = await res.json();
    lastScanData = data;
    renderThumbnails(data);
}

// --- Thumbnails, Selection, Select All/Deselect All ---
const scanResultsContainer = document.getElementById('scan-results-container');
const thumbnailsDiv = document.getElementById('thumbnails');
const selectAllBtn = document.getElementById('select-all');
const deselectAllBtn = document.getElementById('deselect-all');
const createBtn = document.getElementById('create-notes');
const createStatus = document.getElementById('create-status');

function renderThumbnails(notes) {
    if (!Array.isArray(notes) || notes.length === 0) {
        scanResultsContainer.style.display = 'none';
        createBtn.style.display = 'none';
        return;
    }
    scanResultsContainer.style.display = '';
    createBtn.style.display = '';
    thumbnailsDiv.innerHTML = '';
    selectedNotes = notes.map((n, i) => i); // all selected by default
    notes.forEach((note, idx) => {
        const thumb = document.createElement('div');
        thumb.className = 'thumbnail';
        // If note has a color, use as background
        if (note.Color) thumb.style.background = note.Color;
        thumb.innerHTML = `
            <input type="checkbox" class="note-checkbox" data-idx="${idx}" checked>
            <div><strong>${note.Content || 'Note'}</strong></div>
            <div style="font-size:0.8em;">${note.Width}x${note.Height}</div>
        `;
        thumbnailsDiv.appendChild(thumb);
    });
    // Checkbox listeners
    thumbnailsDiv.querySelectorAll('.note-checkbox').forEach(cb => {
        cb.addEventListener('change', (e) => {
            const idx = parseInt(e.target.dataset.idx);
            if (e.target.checked) {
                if (!selectedNotes.includes(idx)) selectedNotes.push(idx);
            } else {
                selectedNotes = selectedNotes.filter(i => i !== idx);
            }
        });
    });
}
selectAllBtn.addEventListener('click', () => {
    thumbnailsDiv.querySelectorAll('.note-checkbox').forEach(cb => { cb.checked = true; });
    if (lastScanData) selectedNotes = lastScanData.map((_, i) => i);
});
deselectAllBtn.addEventListener('click', () => {
    thumbnailsDiv.querySelectorAll('.note-checkbox').forEach(cb => { cb.checked = false; });
    selectedNotes = [];
});

// --- Create Notes ---
createBtn.addEventListener('click', async () => {
    if (!lastScanData || selectedNotes.length === 0) {
        createStatus.textContent = 'Select at least one note.';
        return;
    }
    const notesToSend = selectedNotes.map(i => lastScanData[i]);
    const res = await fetch('/api/create-notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(notesToSend)
    });
    const data = await res.json();
    createStatus.textContent = data.status || data.error || 'Error creating notes.';
});

// --- Camera Capture ---
const captureBtn = document.getElementById('capture-btn');
const cameraStream = document.getElementById('camera-stream');
const captureCanvas = document.getElementById('capture-canvas');
const cameraContainer = document.getElementById('camera-container');
const closeCameraBtn = document.getElementById('close-camera');
let stream = null;
captureBtn.addEventListener('click', async () => {
    if (!stream) {
        try {
            stream = await navigator.mediaDevices.getUserMedia({ video: true });
            cameraStream.srcObject = stream;
            cameraContainer.style.display = 'flex';
            cameraStream.style.display = 'block';
            preview.style.display = 'none';
            imageStatus.textContent = 'Click video to capture photo.';
        } catch (err) {
            imageStatus.textContent = 'Camera access denied.';
            return;
        }
    } else {
        // If already streaming, close camera
        closeCamera();
    }
});
closeCameraBtn.addEventListener('click', closeCamera);
function closeCamera() {
    if (stream) {
        stream.getTracks().forEach(track => track.stop());
        stream = null;
    }
    cameraStream.style.display = 'none';
    cameraContainer.style.display = 'none';
    imageStatus.textContent = '';
}
cameraStream.addEventListener('click', () => {
    if (!stream) return;
    // Draw video frame to canvas
    captureCanvas.width = cameraStream.videoWidth;
    captureCanvas.height = cameraStream.videoHeight;
    captureCanvas.getContext('2d').drawImage(cameraStream, 0, 0);
    // Convert to blob and upload
    captureCanvas.toBlob(async (blob) => {
        preview.src = captureCanvas.toDataURL('image/png');
        preview.style.display = 'block';
        cameraStream.style.display = 'none';
        cameraContainer.style.display = 'none';
        imageStatus.textContent = 'Uploading photo...';
        // Stop camera
        if (stream) {
            stream.getTracks().forEach(track => track.stop());
            stream = null;
        }
        // Upload
        const formData = new FormData();
        formData.append('image', blob, 'capture.png');
        const res = await fetch('/api/upload-image', {
            method: 'POST',
            body: formData
        });
        const data = await res.json();
        imageStatus.textContent = data.status || data.error || 'Upload error.';
        if (data.status === 'uploaded') {
            autoScanNotes();
        }
    }, 'image/png');
});
