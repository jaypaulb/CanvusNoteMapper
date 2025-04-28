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

// --- Fetch Anchors ---
const fetchAnchorsBtn = document.getElementById('fetch-zones');
const canvasSelect = document.getElementById('canvas-select');
const anchorSelect = document.getElementById('zone-select');
const anchorStatus = document.getElementById('zone-status');
fetchAnchorsBtn.addEventListener('click', async () => {
    try {
        const res = await fetch('/api/get-anchors', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({})
        });
        if (!res.ok) {
            const data = await res.json();
            anchorStatus.textContent = data.error || 'Error fetching anchors. Please check your credentials and server connection.';
            return;
        }
        const data = await res.json();
        canvasSelect.innerHTML = '';
        data.canvases.forEach(c => {
            const opt = document.createElement('option');
            opt.value = c.id;
            opt.textContent = c.name;
            canvasSelect.appendChild(opt);
        });
        anchorSelect.innerHTML = '';
        data.anchors.forEach(a => {
            const opt = document.createElement('option');
            opt.value = a.id;
            opt.textContent = a.name;
            anchorSelect.appendChild(opt);
        });
        anchorStatus.textContent = 'Anchors loaded.';
    } catch (err) {
        anchorStatus.textContent = 'Network or server error while fetching canvases/anchors.';
    }
});

// --- Auto-update anchors when canvas changes ---
canvasSelect.addEventListener('change', async () => {
    const selectedCanvas = canvasSelect.value;
    anchorSelect.innerHTML = '';
    anchorStatus.textContent = 'Loading anchors...';
    try {
        const res = await fetch('/api/get-anchors', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ canvasID: selectedCanvas })
        });
        if (!res.ok) {
            const data = await res.json();
            anchorStatus.textContent = data.error || 'Error fetching anchors.';
            return;
        }
        const data = await res.json();
        if (Array.isArray(data.anchors)) {
            anchorData = data.anchors.map(a => ({
                id: a.id,
                name: a.name,
                width: a.width,
                height: a.height,
                x: a.x,
                y: a.y,
                scale: a.scale
            }));
            anchorSelect.innerHTML = '';
            anchorData.forEach(a => {
                const opt = document.createElement('option');
                opt.value = a.id;
                opt.textContent = a.name;
                anchorSelect.appendChild(opt);
            });
            anchorStatus.textContent = 'Anchors loaded.';
        } else {
            anchorStatus.textContent = 'No anchors found for this canvas.';
        }
    } catch (err) {
        anchorStatus.textContent = 'Error fetching anchors.';
    }
    updateCanvasAnchorInfo();
});

// --- Image Capture/Upload & Auto-Scan ---
const imageInput = document.getElementById('image-input');
const imageStatus = document.getElementById('image-status');
const preview = document.getElementById('preview');
let uploadedImage = null;
let lastScanData = null;
let selectedNotes = [];
const imageInputLabel = document.getElementById('image-input-label');
// Make Choose File label trigger file input
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
    // Get selected anchor info
    const anchorOption = anchorSelect.options[anchorSelect.selectedIndex];
    let zoneWidth = 640, zoneHeight = 480, zoneX = 0, zoneY = 0, zoneScale = 1;
    if (anchorOption) {
        const anchor = anchorData.find(a => a.id === anchorOption.value);
        if (anchor) {
            zoneWidth = Math.round(anchor.width);
            zoneHeight = Math.round(anchor.height);
            zoneX = Math.round(anchor.x);
            zoneY = Math.round(anchor.y);
            zoneScale = anchor.scale !== undefined ? anchor.scale : 1;
        }
    }
    const res = await fetch('/api/scan-notes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            imageData: [],
            imageDimensions: [640, 480],
            zoneDimensions: [zoneWidth, zoneHeight],
            zoneLocation: [zoneX, zoneY],
            zoneScale: zoneScale
        })
    });
    const data = await res.json();
    lastScanData = data;
    renderThumbnails(data.notes || []);
}

// --- Thumbnails, Selection, Select All/Deselect All ---
const scanResultsContainer = document.getElementById('scan-results-container');
const thumbnailsDiv = document.getElementById('thumbnails');
const selectAllBtn = document.getElementById('select-all');
const deselectAllBtn = document.getElementById('deselect-all');
const createBtn = document.getElementById('create-notes');
const createStatus = document.getElementById('create-status');
let anchorData = [];

// Add display for selected canvas and anchor info
let canvasAnchorInfo = document.getElementById('canvas-anchor-info');
if (!canvasAnchorInfo) {
    canvasAnchorInfo = document.createElement('div');
    canvasAnchorInfo.id = 'canvas-anchor-info';
    canvasAnchorInfo.style.margin = '0.5em 0';
    createBtn.parentNode.insertBefore(canvasAnchorInfo, createBtn);
}

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
        // Use MCS API fields
        if (note.background_color) thumb.style.background = note.background_color;
        thumb.innerHTML = `
            <input type="checkbox" class="note-checkbox" data-idx="${idx}" checked>
            <div><strong>${note.text || 'Note'}</strong></div>
            <div style="font-size:0.8em;">${note.size?.width || 0}x${note.size?.height || 0}</div>
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
    updateCanvasAnchorInfo();
}

// --- Canvas/Zone selection display ---
function updateCanvasAnchorInfo() {
    const canvasName = canvasSelect.options[canvasSelect.selectedIndex]?.text || '';
    const anchorOption = anchorSelect.options[anchorSelect.selectedIndex];
    let anchorInfo = '';
    if (anchorOption) {
        const anchor = anchorData.find(a => a.id === anchorOption.value);
        if (anchor) {
            const x = Math.floor(anchor.x);
            const y = Math.floor(anchor.y);
            anchorInfo = ` → ${anchor.name} [${anchor.width}x${anchor.height} @ ${x}, ${y}]`;
        }
    }
    canvasAnchorInfo.textContent = canvasName + (anchorInfo ? anchorInfo : '');
}

canvasSelect.addEventListener('change', updateCanvasAnchorInfo);
anchorSelect.addEventListener('change', updateCanvasAnchorInfo);

// --- Camera Capture ---
const captureBtn = document.getElementById('capture-btn');
const cameraStream = document.getElementById('camera-stream');
const captureCanvas = document.getElementById('capture-canvas');
const cameraContainer = document.getElementById('camera-container');
const closeCameraBtn = document.getElementById('close-camera');
let stream = null;

// Toggle camera on/off with Capture button
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
        closeCamera();
    }
});

// Always show close button when camera is open
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

// Capture image from video stream
cameraStream.addEventListener('click', () => {
    if (!stream) return;
    captureCanvas.width = cameraStream.videoWidth;
    captureCanvas.height = cameraStream.videoHeight;
    captureCanvas.getContext('2d').drawImage(cameraStream, 0, 0);
    captureCanvas.toBlob(async (blob) => {
        preview.src = captureCanvas.toDataURL('image/png');
        preview.style.display = 'block';
        preview.style.margin = '1em auto 0 auto'; // Center preview
        cameraStream.style.display = 'none';
        cameraContainer.style.display = 'none';
        imageStatus.textContent = 'Uploading photo...';
        if (stream) {
            stream.getTracks().forEach(track => track.stop());
            stream = null;
        }
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

// --- Create Notes ---
createBtn.addEventListener('click', async () => {
    if (!lastScanData || selectedNotes.length === 0) {
        createStatus.textContent = 'Select at least one note.';
        return;
    }
    // Get selected canvas and zone
    const canvasID = canvasSelect.value;
    const zoneID = anchorSelect.value;
    if (!canvasID || !zoneID) {
        createStatus.textContent = 'Please select both a canvas and a zone before creating notes.';
        return;
    }
    const notesToSend = selectedNotes.map(i => (lastScanData.notes ? lastScanData.notes[i] : lastScanData[i]));
    try {
        const res = await fetch('/api/create-notes', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ canvasID, zoneID, notes: notesToSend })
        });
        const data = await res.json();
        if (res.ok && data.status) {
            createStatus.textContent = data.status;
        } else {
            createStatus.textContent = data.error || 'Error creating notes. Please check your canvas/zone selection and try again.';
        }
    } catch (err) {
        createStatus.textContent = 'Network or server error while creating notes.';
    }
});
