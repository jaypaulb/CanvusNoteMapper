window.addEventListener('DOMContentLoaded', () => {
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
    const mcsServer = localStorage.getItem('mcsServer');
    const apiKey = localStorage.getItem('apiKey');
    if (mcsServer) document.getElementById('mcs-server').value = mcsServer;
    if (apiKey) document.getElementById('api-key').value = apiKey;

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
        if (data.status === 'ok') {
            await fetchCanvases();
            // Switch to Scan tab
            const scanTabBtn = document.getElementById('tab-scan');
            if (scanTabBtn) {
                scanTabBtn.classList.add('active');
                // Remove active from other tab buttons
                tabBtns.forEach(b => { if (b !== scanTabBtn) b.classList.remove('active'); });
                // Hide all tab sections
                tabSections.forEach(sec => sec.style.display = 'none');
                // Show scan section
                document.getElementById('scan-section').style.display = '';
            }
        }
    });

    // --- Fetch Canvases and Anchors ---
    const canvasSelect = document.getElementById('canvas-select');
    const anchorSelect = document.getElementById('zone-select');
    const anchorStatus = document.getElementById('zone-status');
    const canvasAnchorInfo = document.getElementById('canvas-anchor-info');

    async function fetchCanvases() {
        try {
            const res = await fetch('/api/get-canvases');
            if (!res.ok) {
                let data = {};
                try { data = await res.json(); } catch (e) {}
                const msg = data.error || 'Error fetching canvases. Please check your credentials or server connection.';
                anchorStatus.textContent = msg;
                console.error('[fetchCanvases] Server error:', msg);
                return;
            }
            const data = await res.json();
            canvasSelect.innerHTML = '';
            if (data.canvases) {
                data.canvases.forEach(c => {
                    const opt = document.createElement('option');
                    opt.value = c.id;
                    opt.textContent = c.name;
                    canvasSelect.appendChild(opt);
                });
            }
            if (canvasSelect.options.length > 0) {
                await fetchAnchors(canvasSelect.value);
            } else {
                anchorStatus.textContent = 'No canvases found. Please check your credentials or server connection.';
                console.warn('[fetchCanvases] No canvases found in response.');
            }
        } catch (err) {
            anchorStatus.textContent = 'Network or server error while fetching canvases. Please check your connection or server logs.';
            console.error('[fetchCanvases] Network or server error:', err);
            setTimeout(() => { anchorStatus.textContent = ''; }, 5000);
        }
    }

    async function fetchAnchors(canvasID) {
        anchorSelect.innerHTML = '';
        anchorStatus.textContent = 'Loading anchors...';
        try {
            const res = await fetch(`/api/get-anchors?canvasID=${encodeURIComponent(canvasID)}`);
            if (!res.ok) {
                const data = await res.json();
                anchorStatus.textContent = data.error || 'Error fetching anchors.';
                return;
            }
            const data = await res.json();
            if (Array.isArray(data.anchors)) {
                anchorData = data.anchors.map(a => ({
                    id: a.id,
                    name: a.name
                }));
                anchorSelect.innerHTML = '';
                anchorData.forEach(a => {
                    const opt = document.createElement('option');
                    opt.value = a.id;
                    opt.textContent = a.name;
                    anchorSelect.appendChild(opt);
                });
                anchorStatus.textContent = '';
            } else {
                anchorStatus.textContent = 'No anchors found for this canvas.';
            }
        } catch (err) {
            anchorStatus.textContent = 'Error fetching anchors.';
        }
        updateCanvasAnchorInfo();
    }

    async function fetchAnchorInfo(canvasID, anchorID) {
        anchorStatus.textContent = 'Refreshing anchor details...';
        try {
            const url = `/api/get-anchor-info?canvasID=${encodeURIComponent(canvasID)}&anchorID=${encodeURIComponent(anchorID)}`;
            const res = await fetch(url);
            if (!res.ok) {
                let errorText = '';
                try { errorText = await res.text(); } catch (e) {}
                console.error('[fetchAnchorInfo] Error response:', errorText);
                anchorStatus.textContent = 'Error fetching anchor details.';
                return null;
            }
            const anchor = await res.json();
            if (!anchor.id || !anchor.name) {
                console.error('Anchor details missing required fields:', anchor);
                anchorStatus.textContent = 'Anchor details missing required fields.';
                return null;
            }
            anchorStatus.textContent = '';
            return anchor;
        } catch (err) {
            console.error('[fetchAnchorInfo] Network or server error:', err);
            anchorStatus.textContent = 'Network or server error while fetching anchor details.';
            return null;
        }
    }

    // updateCanvasAnchorInfo now fetches anchor info for the selected anchor
    async function updateCanvasAnchorInfo() {
        const anchorOption = anchorSelect.options[anchorSelect.selectedIndex];
        if (!anchorOption) return;
        const canvasID = canvasSelect.value;
        const anchorID = anchorOption.value;
        const anchor = await fetchAnchorInfo(canvasID, anchorID);
        if (anchor) {
            const x = Math.floor(anchor.x);
            const y = Math.floor(anchor.y);
            anchorOption.textContent = `${anchor.name} [${anchor.width}x${anchor.height} @ ${x}, ${y}]`;
        } else {
            anchorOption.textContent = anchorOption.value;
        }
    }

    if (canvasSelect) {
        canvasSelect.addEventListener('change', async () => {
            await fetchAnchors(canvasSelect.value); // repopulate anchor dropdown
            if (anchorSelect.options.length > 0) {
                anchorSelect.selectedIndex = 0; // select first anchor
                await updateCanvasAnchorInfo(); // fetch info for new anchor
            } else {
                // No anchors for this canvas, clear anchor info display
                // Optionally, set anchorStatus or clear anchorSelect
                anchorStatus.textContent = 'No anchors found for this canvas.';
            }
        });
    }

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
                zoneScale = anchor.scale !== undefined ? Math.round(anchor.scale * 100) / 100 : 1;
            }
        }
        imageStatus.textContent = 'Processing image with AI...';
        const formData = new FormData();
        formData.append('image', uploadedImage);
        formData.append('zoneDimensions', JSON.stringify([zoneWidth, zoneHeight]));
        formData.append('zoneLocation', JSON.stringify([zoneX, zoneY]));
        formData.append('zoneScale', JSON.stringify(zoneScale));
        const res = await fetch('/api/scan-notes', {
            method: 'POST',
            body: formData
        });
        const data = await res.json();
        lastScanData = data;
        renderThumbnails(data.notes || []);
        imageStatus.textContent = '';
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
            if (scanResultsContainer) scanResultsContainer.style.display = 'none';
            if (createBtn) createBtn.style.display = 'none';
            return;
        }
        if (scanResultsContainer) scanResultsContainer.style.display = '';
        if (createBtn) createBtn.style.display = '';
        if (thumbnailsDiv) thumbnailsDiv.innerHTML = '';
        // If selectedNotes is empty, select all by default
        if (!selectedNotes || selectedNotes.length === 0) {
            selectedNotes = notes.map((n, i) => i);
        }
        notes.forEach((note, idx) => {
            const thumb = document.createElement('div');
            thumb.className = 'thumbnail';
            if (note.background_color) thumb.style.background = note.background_color;
            thumb.innerHTML = `
                <input type="checkbox" class="note-checkbox" data-idx="${idx}" ${selectedNotes.includes(idx) ? 'checked' : ''}>
                <div><strong>${note.text || 'Note'}</strong></div>
                <div style="font-size:0.8em;">${note.size?.width || 0}x${note.size?.height || 0}</div>
            `;
            if (thumbnailsDiv) thumbnailsDiv.appendChild(thumb);
        });
        // Checkbox listeners
        if (thumbnailsDiv) {
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
        updateCanvasAnchorInfo();
    }

    if (selectAllBtn) {
        selectAllBtn.addEventListener('click', () => {
            if (!thumbnailsDiv) return;
            const checkboxes = thumbnailsDiv.querySelectorAll('.note-checkbox');
            selectedNotes = (lastScanData?.notes || []).map((_, i) => i);
            checkboxes.forEach(cb => { cb.checked = true; });
        });
    } else {
        console.error('selectAllBtn not found in DOM!');
    }

    if (deselectAllBtn) {
        deselectAllBtn.addEventListener('click', () => {
            if (!thumbnailsDiv) return;
            const checkboxes = thumbnailsDiv.querySelectorAll('.note-checkbox');
            selectedNotes = [];
            checkboxes.forEach(cb => { cb.checked = false; });
        });
    } else {
        console.error('deselectAllBtn not found in DOM!');
    }

    // --- Camera Capture ---
    const captureBtn = document.getElementById('capture-btn');
    const cameraStream = document.getElementById('camera-stream');
    const captureCanvas = document.getElementById('capture-canvas');
    const cameraContainer = document.getElementById('camera-container');
    const closeCameraBtn = document.getElementById('close-camera');
    const switchCameraBtn = document.getElementById('switch-camera');
    let stream = null;
    let currentFacingMode = 'environment'; // default to rear camera
    let videoInputDevices = [];
    let currentDeviceIndex = 0;

    async function getVideoInputDevices() {
        const devices = await navigator.mediaDevices.enumerateDevices();
        return devices.filter(device => device.kind === 'videoinput');
    }

    async function startCamera(constraints) {
        if (stream) {
            stream.getTracks().forEach(track => track.stop());
        }
        try {
            stream = await navigator.mediaDevices.getUserMedia(constraints);
            cameraStream.srcObject = stream;
            cameraContainer.style.display = 'flex';
            cameraStream.style.display = 'block';
            preview.style.display = 'none';
            imageStatus.textContent = 'Click video to capture photo.';
        } catch (err) {
            imageStatus.textContent = 'Camera access denied.';
            stream = null;
        }
    }

    if (captureBtn) {
        captureBtn.addEventListener('click', async () => {
            if (!stream) {
                videoInputDevices = await getVideoInputDevices();
                if (videoInputDevices.length > 0) {
                    // Try to use facingMode if supported, else fallback to deviceId
                    let constraints = { video: { facingMode: currentFacingMode } };
                    // If only one camera, fallback to deviceId
                    if (videoInputDevices.length === 1) {
                        constraints = { video: { deviceId: { exact: videoInputDevices[0].deviceId } } };
                    }
                    await startCamera(constraints);
                } else {
                    imageStatus.textContent = 'No camera devices found.';
                }
            } else {
                closeCamera();
            }
        });
    } else {
        console.error('captureBtn not found in DOM!');
    }

    if (switchCameraBtn) {
        switchCameraBtn.addEventListener('click', async () => {
            videoInputDevices = await getVideoInputDevices();
            if (videoInputDevices.length > 1) {
                // Cycle to next camera
                currentDeviceIndex = (currentDeviceIndex + 1) % videoInputDevices.length;
                const deviceId = videoInputDevices[currentDeviceIndex].deviceId;
                await startCamera({ video: { deviceId: { exact: deviceId } } });
            } else {
                // Toggle facingMode if only one camera (may not always work)
                currentFacingMode = currentFacingMode === 'environment' ? 'user' : 'environment';
                await startCamera({ video: { facingMode: currentFacingMode } });
            }
        });
    }

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
    async function createNotes() {
        if (!lastScanData || selectedNotes.length === 0) {
            createStatus.textContent = 'Select at least one note.';
            return;
        }
        const canvasID = canvasSelect.value;
        const zoneID = anchorSelect.value;
        if (!canvasID || !zoneID) {
            createStatus.textContent = 'Please select both a canvas and a zone before creating notes.';
            return;
        }
        const anchor = await fetchAnchorInfo(canvasID, zoneID);
        if (!anchor) {
            createStatus.textContent = 'Failed to refresh anchor details.';
            return;
        }
        // Send notes as detected (raw), backend will handle transformation
        const notesToSend = selectedNotes.map(i => (lastScanData.notes ? lastScanData.notes[i] : lastScanData[i]));
        try {
            const res = await fetch('/api/create-notes', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    canvasID,
                    zoneID,
                    notes: notesToSend,
                    imageWidth: lastScanData.imageWidth,
                    imageHeight: lastScanData.imageHeight
                })
            });
            const data = await res.json();
            if (res.ok && data.status) {
                createStatus.textContent = data.status;
            } else {
                createStatus.textContent = data.error || 'Error creating notes. Please check your canvas/zone selection and try again.';
            }
        } catch (err) {
            console.error('[Create Notes] Network or server error:', err);
            createStatus.textContent = 'Network or server error while creating notes.';
        }
    }
    if (!createBtn._bound) {
        createBtn.addEventListener('click', createNotes);
        createBtn._bound = true;
    }
});
