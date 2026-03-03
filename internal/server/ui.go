package server

// uiHTML is the embedded single-page web UI.
const uiHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>video2gif — Production GIF Converter</title>
<style>
  @import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&family=Space+Grotesk:wght@300;400;500;700&display=swap');

  :root {
    --bg: #0a0a0f;
    --surface: #111118;
    --surface2: #1a1a26;
    --border: #2a2a3e;
    --accent: #7c3aed;
    --accent2: #a78bfa;
    --accent3: #34d399;
    --text: #e2e8f0;
    --muted: #64748b;
    --danger: #ef4444;
    --warn: #f59e0b;
  }

  * { margin: 0; padding: 0; box-sizing: border-box; }

  body {
    font-family: 'Space Grotesk', sans-serif;
    background: var(--bg);
    color: var(--text);
    min-height: 100vh;
    overflow-x: hidden;
  }

  /* Animated grid background */
  body::before {
    content: '';
    position: fixed; inset: 0;
    background-image:
      linear-gradient(rgba(124,58,237,0.04) 1px, transparent 1px),
      linear-gradient(90deg, rgba(124,58,237,0.04) 1px, transparent 1px);
    background-size: 40px 40px;
    pointer-events: none;
    z-index: 0;
  }

  .container { max-width: 1100px; margin: 0 auto; padding: 0 24px; position: relative; z-index: 1; }

  header {
    padding: 32px 0 24px;
    border-bottom: 1px solid var(--border);
  }

  .logo {
    font-family: 'JetBrains Mono', monospace;
    font-size: 1.6rem;
    font-weight: 700;
    letter-spacing: -0.03em;
    background: linear-gradient(135deg, var(--accent2), var(--accent3));
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .logo span { color: var(--muted); -webkit-text-fill-color: var(--muted); font-weight: 400; }

  .subtitle { color: var(--muted); font-size: 0.85rem; margin-top: 4px; }

  .health-dot {
    display: inline-block; width: 8px; height: 8px;
    border-radius: 50%; background: var(--accent3);
    box-shadow: 0 0 8px var(--accent3);
    margin-right: 6px; animation: pulse 2s infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; } 50% { opacity: 0.4; }
  }

  main { padding: 32px 0; display: grid; grid-template-columns: 1fr 380px; gap: 24px; }

  @media (max-width: 800px) { main { grid-template-columns: 1fr; } }

  /* Card */
  .card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 24px;
  }

  .card-title {
    font-size: 0.7rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.15em;
    color: var(--muted);
    margin-bottom: 20px;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .card-title::after {
    content: '';
    flex: 1;
    height: 1px;
    background: var(--border);
  }

  /* Drop zone */
  .dropzone {
    border: 2px dashed var(--border);
    border-radius: 10px;
    padding: 48px 24px;
    text-align: center;
    cursor: pointer;
    transition: all 0.2s;
    position: relative;
    background: rgba(124,58,237,0.02);
  }

  .dropzone:hover, .dropzone.dragover {
    border-color: var(--accent);
    background: rgba(124,58,237,0.08);
  }

  .dropzone input[type=file] {
    position: absolute; inset: 0; opacity: 0; cursor: pointer; width: 100%; height: 100%;
  }

  .drop-icon { font-size: 2.5rem; margin-bottom: 12px; }

  .drop-text { color: var(--muted); font-size: 0.9rem; }
  .drop-text strong { color: var(--text); }
  .drop-formats { font-size: 0.75rem; color: var(--muted); margin-top: 6px; font-family: 'JetBrains Mono', monospace; }

  .recorder-row {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-bottom: 10px;
  }

  .recorder-status {
    margin-left: auto;
    font-size: 0.75rem;
    color: var(--muted);
    font-family: 'JetBrains Mono', monospace;
  }

  .recorder-status.recording {
    color: #f87171;
  }

  /* File preview */
  .file-preview {
    display: none;
    align-items: center;
    gap: 12px;
    padding: 12px 14px;
    background: var(--surface2);
    border-radius: 8px;
    margin-top: 14px;
  }

  .file-preview .file-icon { font-size: 1.4rem; }
  .file-preview .file-info { flex: 1; min-width: 0; }
  .file-preview .file-name { font-size: 0.85rem; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .file-preview .file-meta { font-size: 0.75rem; color: var(--muted); font-family: 'JetBrains Mono', monospace; }
  .file-preview .remove-btn { cursor: pointer; color: var(--danger); font-size: 1.1rem; }

  .editor-wrap {
    display: none;
    margin-top: 16px;
    padding: 14px;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--surface2);
  }

  .editor-video {
    width: 100%;
    max-height: 280px;
    border-radius: 8px;
    background: #000;
  }

  .editor-toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 10px;
    align-items: center;
  }

  .chip-btn {
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text);
    border-radius: 6px;
    padding: 6px 10px;
    font-size: 0.72rem;
    cursor: pointer;
    font-family: 'JetBrains Mono', monospace;
  }

  .chip-btn:hover { border-color: var(--accent2); }
  .chip-btn.warn { color: var(--warn); border-color: rgba(245,158,11,0.5); }
  .chip-btn.danger { color: var(--danger); border-color: rgba(239,68,68,0.5); }

  .editor-time {
    margin-left: auto;
    font-size: 0.74rem;
    color: var(--accent2);
    font-family: 'JetBrains Mono', monospace;
  }

  .timeline-wrap { margin-top: 12px; }
  .timeline-track {
    width: 100%;
    height: 16px;
    border-radius: 999px;
    background: #0b0f16;
    border: 1px solid var(--border);
    position: relative;
    overflow: hidden;
  }

  .timeline-seg {
    position: absolute;
    top: 0;
    height: 100%;
    background: linear-gradient(90deg, #34d399, #10b981);
    box-shadow: inset 0 0 0 1px rgba(16,185,129,0.65);
    pointer-events: none;
    opacity: 0.6;
  }

  .timeline-cut {
    position: absolute;
    top: 0;
    height: 100%;
    background: linear-gradient(90deg, rgba(239,68,68,0.55), rgba(248,113,113,0.55));
    box-shadow: inset 0 0 0 1px rgba(239,68,68,0.85);
    border-radius: 999px;
  }

  .timeline-handle {
    position: absolute;
    top: -2px;
    width: 10px;
    height: 20px;
    border-radius: 999px;
    background: #fca5a5;
    border: 1px solid rgba(239,68,68,0.95);
    cursor: ew-resize;
  }

  .timeline-handle.left { left: -5px; }
  .timeline-handle.right { right: -5px; }

  .timeline-cut-label {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.64rem;
    color: #fee2e2;
    font-family: 'JetBrains Mono', monospace;
    pointer-events: none;
  }

  .timeline-playhead {
    position: absolute;
    top: -3px;
    width: 2px;
    height: 22px;
    background: var(--accent2);
    box-shadow: 0 0 8px rgba(167,139,250,0.6);
  }

  .timeline-scrub {
    margin-top: 8px;
    width: 100%;
  }

  .cut-controls {
    margin-top: 10px;
    display: grid;
    grid-template-columns: 1fr 1fr auto;
    gap: 8px;
    align-items: end;
  }

  .cut-controls input[type=number] {
    min-width: 0;
  }

  .segment-list {
    margin-top: 10px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .segment-title {
    margin-top: 10px;
    font-size: 0.72rem;
    color: var(--muted);
    text-transform: uppercase;
    letter-spacing: 0.08em;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }

  .segment-row {
    display: grid;
    grid-template-columns: 42px 1fr 1fr auto auto auto auto;
    gap: 6px;
    align-items: center;
  }

  .segment-row-id {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.72rem;
    color: #fca5a5;
  }

  .segment-input {
    background: #0f1420;
    border: 1px solid rgba(239,68,68,0.35);
    color: #fecaca;
    border-radius: 6px;
    padding: 6px 8px;
    font-size: 0.74rem;
    font-family: 'JetBrains Mono', monospace;
    width: 100%;
  }

  .segment-input:focus {
    outline: none;
    border-color: rgba(239,68,68,0.8);
  }

  .segment-pill {
    border: 1px solid rgba(239,68,68,0.45);
    background: rgba(239,68,68,0.12);
    color: #fca5a5;
    border-radius: 999px;
    padding: 5px 6px 5px 9px;
    font-size: 0.7rem;
    font-family: 'JetBrains Mono', monospace;
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  .segment-pill-btn {
    border: 1px solid rgba(239,68,68,0.55);
    background: rgba(239,68,68,0.2);
    color: #fecaca;
    border-radius: 999px;
    width: 18px;
    height: 18px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    font-size: 0.7rem;
    line-height: 1;
    padding: 0;
  }

  .segment-empty {
    color: var(--muted);
    font-size: 0.76rem;
    font-family: 'JetBrains Mono', monospace;
    padding: 2px 0;
  }

  /* Profile selector */
  .profiles-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 8px;
    margin-bottom: 20px;
  }

  .profile-btn {
    padding: 10px 8px;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: transparent;
    color: var(--text);
    cursor: pointer;
    font-family: 'Space Grotesk', sans-serif;
    font-size: 0.78rem;
    font-weight: 500;
    transition: all 0.15s;
    text-align: center;
  }

  .profile-btn:hover { border-color: var(--accent2); background: rgba(124,58,237,0.1); }

  .profile-btn.active {
    border-color: var(--accent);
    background: rgba(124,58,237,0.2);
    color: var(--accent2);
  }

  .profile-desc { font-size: 0.7rem; color: var(--muted); margin-top: 3px; }

  /* Parameters grid */
  .params-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }

  .param-group { display: flex; flex-direction: column; gap: 5px; }

  .param-group.full { grid-column: 1 / -1; }

  label {
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--muted);
  }

  input[type=number], input[type=text], select {
    background: var(--surface2);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 8px 10px;
    color: var(--text);
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.82rem;
    width: 100%;
    transition: border-color 0.15s;
  }

  input:focus, select:focus {
    outline: none;
    border-color: var(--accent);
  }

  /* Range slider */
  .slider-row { display: flex; align-items: center; gap: 10px; }

  input[type=range] {
    flex: 1;
    accent-color: var(--accent);
    height: 4px;
    cursor: pointer;
    background: transparent;
    padding: 0;
    border: none;
  }

  .slider-val {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.8rem;
    color: var(--accent2);
    min-width: 35px;
    text-align: right;
  }

  /* Convert button */
  .convert-btn {
    width: 100%;
    padding: 14px;
    margin-top: 20px;
    background: linear-gradient(135deg, var(--accent), #5b21b6);
    border: none;
    border-radius: 8px;
    color: #fff;
    font-family: 'Space Grotesk', sans-serif;
    font-size: 0.95rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    letter-spacing: 0.02em;
    position: relative;
    overflow: hidden;
  }

  .convert-btn:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 8px 24px rgba(124,58,237,0.4);
  }

  .convert-btn:disabled { opacity: 0.5; cursor: not-allowed; transform: none; }

  .action-row {
    margin-top: 20px;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }

  @media (max-width: 800px) {
    .action-row {
      grid-template-columns: 1fr;
    }
  }

  /* Progress */
  .progress-wrap {
    display: none;
    margin-top: 16px;
    background: var(--surface2);
    border-radius: 8px;
    padding: 14px;
  }

  .progress-label { font-size: 0.78rem; color: var(--muted); margin-bottom: 8px; display: flex; justify-content: space-between; }

  .progress-bar {
    height: 4px;
    background: var(--border);
    border-radius: 2px;
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    background: linear-gradient(90deg, var(--accent), var(--accent3));
    border-radius: 2px;
    transition: width 0.4s;
    width: 0%;
  }

  .progress-status { font-size: 0.75rem; color: var(--accent2); margin-top: 6px; font-family: 'JetBrains Mono', monospace; }

  /* Jobs list */
  .jobs-list { display: flex; flex-direction: column; gap: 10px; }

  .job-item {
    background: var(--surface2);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 12px 14px;
    display: flex;
    align-items: center;
    gap: 12px;
    transition: border-color 0.2s;
  }

  .job-item:hover { border-color: var(--accent2); }

  .job-status-dot {
    width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0;
  }

  .status-queued  { background: var(--warn); }
  .status-running { background: var(--accent2); animation: pulse 1s infinite; }
  .status-done    { background: var(--accent3); }
  .status-failed  { background: var(--danger); }

  .job-info { flex: 1; min-width: 0; }
  .job-name { font-size: 0.82rem; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .job-meta { font-size: 0.7rem; color: var(--muted); font-family: 'JetBrains Mono', monospace; margin-top: 2px; }

  .job-actions { display: flex; gap: 6px; }

  .btn-sm {
    padding: 5px 10px;
    border-radius: 6px;
    font-size: 0.72rem;
    font-weight: 600;
    cursor: pointer;
    border: 1px solid;
    font-family: 'Space Grotesk', sans-serif;
    transition: all 0.15s;
  }

  .btn-download {
    border-color: var(--accent3);
    color: var(--accent3);
    background: transparent;
  }

  .btn-download:hover { background: rgba(52,211,153,0.15); }

  .btn-delete {
    border-color: var(--border);
    color: var(--muted);
    background: transparent;
  }

  .btn-delete:hover { border-color: var(--danger); color: var(--danger); }

  .empty-state {
    text-align: center;
    color: var(--muted);
    padding: 32px;
    font-size: 0.85rem;
  }

  /* Config panel */
  .config-panel { margin-top: 16px; }

  .config-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 0;
    border-bottom: 1px solid var(--border);
    font-size: 0.82rem;
  }

  .config-row:last-child { border-bottom: none; }
  .config-key { color: var(--muted); font-family: 'JetBrains Mono', monospace; font-size: 0.75rem; }
  .config-val { color: var(--accent2); font-family: 'JetBrains Mono', monospace; font-size: 0.75rem; }

  /* Toast */
  #toast {
    position: fixed;
    bottom: 24px; right: 24px;
    padding: 12px 20px;
    border-radius: 8px;
    font-size: 0.85rem;
    font-weight: 500;
    z-index: 9999;
    transform: translateY(60px);
    opacity: 0;
    transition: all 0.3s;
    max-width: 320px;
  }

  #toast.show { transform: translateY(0); opacity: 1; }
  #toast.success { background: rgba(52,211,153,0.15); border: 1px solid var(--accent3); color: var(--accent3); }
  #toast.error { background: rgba(239,68,68,0.15); border: 1px solid var(--danger); color: var(--danger); }

  .tag {
    display: inline-block;
    padding: 2px 7px;
    border-radius: 4px;
    font-size: 0.65rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    background: rgba(124,58,237,0.2);
    color: var(--accent2);
    border: 1px solid rgba(124,58,237,0.3);
  }
</style>
</head>
<body>
<div class="container">
  <header>
    <div style="display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:12px">
      <div>
        <div class="logo">video<span>2</span>gif</div>
        <div class="subtitle"><span class="health-dot" id="healthDot"></span>Production GIF Converter — FFmpeg Powered</div>
      </div>
      <div style="display:flex;gap:10px;align-items:center">
        <span id="statsLabel" style="font-size:0.75rem;color:var(--muted);font-family:'JetBrains Mono',monospace"></span>
        <span class="tag">v1.0.0</span>
      </div>
    </div>
  </header>

  <main>
    <!-- Left: Convert panel -->
    <div>
      <div class="card">
        <div class="card-title">Upload &amp; Configure</div>

        <div class="recorder-row">
          <button class="chip-btn" type="button" id="recordStartBtn" onclick="startScreenRecording()">Start Recording</button>
          <button class="chip-btn danger" type="button" id="recordStopBtn" onclick="stopScreenRecording()" disabled>Stop Recording</button>
          <span class="recorder-status" id="recorderStatus">Recorder idle</span>
        </div>

        <div class="dropzone" id="dropzone">
          <input type="file" id="fileInput" accept="video/*,.mkv,.avi,.flv,.wmv,.ts,.mts,.m2ts" />
          <div class="drop-icon">🎬</div>
          <div class="drop-text"><strong>Drop a video file here</strong> or click to browse</div>
          <div class="drop-formats">MP4 · MOV · MKV · AVI · WEBM · FLV · WMV · TS · 3GP</div>
        </div>

        <div class="file-preview" id="filePreview">
          <div class="file-icon">📹</div>
          <div class="file-info">
            <div class="file-name" id="previewName"></div>
            <div class="file-meta" id="previewMeta"></div>
          </div>
          <div class="remove-btn" onclick="clearFile()" title="Remove">✕</div>
        </div>

        <div class="editor-wrap" id="editorWrap">
          <video id="editorVideo" class="editor-video" controls preload="metadata"></video>
          <div class="editor-toolbar">
            <button class="chip-btn" type="button" onclick="markCutStart()">Mark Cut Start</button>
            <button class="chip-btn" type="button" onclick="markCutEnd()">Mark Cut End</button>
            <button class="chip-btn warn" type="button" onclick="cutMarkedRange()">Cut Marked Range</button>
            <button class="chip-btn danger" type="button" onclick="resetCuts()">Reset Cuts</button>
            <span class="editor-time" id="editorNow">00:00.00 / 00:00.00</span>
          </div>

          <div class="timeline-wrap">
            <div class="timeline-track" id="timelineTrack">
              <div class="timeline-playhead" id="timelinePlayhead"></div>
            </div>
            <input class="timeline-scrub" type="range" id="timelineScrub" min="0" max="0" step="0.01" value="0">
          </div>

          <div class="cut-controls">
            <div class="param-group">
              <label>Cut Start (s)</label>
              <input type="number" id="cutStart" min="0" step="0.01" value="0">
            </div>
            <div class="param-group">
              <label>Cut End (s)</label>
              <input type="number" id="cutEnd" min="0" step="0.01" value="0">
            </div>
            <button class="chip-btn warn" type="button" onclick="addCutFromInputs()">Add Cut Range</button>
          </div>

          <div class="segment-title">
            <span>Cut Ranges (Multi-Select)</span>
            <button class="chip-btn danger" type="button" onclick="resetCuts()">Reset All</button>
          </div>
          <div class="segment-list" id="segmentList"></div>
        </div>
      </div>

      <div class="card" style="margin-top:16px">
        <div class="card-title">Quality Profile</div>

        <div class="profiles-grid" id="profilesGrid"></div>

        <div class="card-title" style="margin-top:20px">Fine-tune Parameters</div>

        <div class="params-grid">
          <div class="param-group">
            <label>FPS</label>
            <div class="slider-row">
              <input type="range" id="fps" min="1" max="60" step="0.5" value="20" oninput="syncSlider('fps','fpsVal')">
              <span class="slider-val" id="fpsVal">20</span>
            </div>
          </div>

          <div class="param-group">
            <label>Colors (Palette)</label>
            <div class="slider-row">
              <input type="range" id="colors" min="2" max="256" step="2" value="256" oninput="syncSlider('colors','colorsVal')">
              <span class="slider-val" id="colorsVal">256</span>
            </div>
          </div>

          <div class="param-group">
            <label>Width (px, -1=auto)</label>
            <input type="number" id="width" value="640" min="-1" max="4096" step="2">
          </div>

          <div class="param-group">
            <label>Height (px, -1=auto)</label>
            <input type="number" id="height" value="-1" min="-1" max="4096" step="2">
          </div>

          <div class="param-group">
            <label>Dither Algorithm</label>
            <select id="dither">
              <option value="sierra2_4a">sierra2_4a (balanced)</option>
              <option value="sierra2">sierra2 (quality)</option>
              <option value="bayer">bayer (fast)</option>
              <option value="floyd_steinberg">floyd_steinberg</option>
              <option value="none">none (smallest)</option>
            </select>
          </div>

          <div class="param-group">
            <label>Bayer Scale (0–5)</label>
            <div class="slider-row">
              <input type="range" id="bayerScale" min="0" max="5" step="1" value="2" oninput="syncSlider('bayerScale','bayerVal')">
              <span class="slider-val" id="bayerVal">2</span>
            </div>
          </div>

          <div class="param-group">
            <label>Speed Multiplier</label>
            <div class="slider-row">
              <input type="range" id="speed" min="0.25" max="4" step="0.25" value="1" oninput="syncSlider('speed','speedVal')">
              <span class="slider-val" id="speedVal">1×</span>
            </div>
          </div>

          <div class="param-group">
            <label>Loop (0=∞, -1=no loop)</label>
            <input type="number" id="loop" value="0" min="-1" max="100">
          </div>

          <div class="param-group full" style="flex-direction:row;align-items:center;gap:10px">
            <input type="checkbox" id="optimizePalette" checked style="accent-color:var(--accent);width:16px;height:16px">
            <label for="optimizePalette" style="text-transform:none;font-size:0.82rem;color:var(--text);letter-spacing:0">
              Two-pass palette optimization (better quality, slower)
            </label>
          </div>
        </div>

        <div class="action-row">
          <button class="convert-btn" id="convertBtn" onclick="startConvert()" disabled>
            Convert to GIF
          </button>
          <button class="convert-btn" id="saveBtn" onclick="saveEditedVideo()" disabled>
            Save Edited Video
          </button>
        </div>

        <div class="progress-wrap" id="progressWrap">
          <div class="progress-label">
            <span id="progressStatus">Processing...</span>
            <span id="progressPct">—</span>
          </div>
          <div class="progress-bar"><div class="progress-fill" id="progressFill"></div></div>
          <div class="progress-status" id="progressDetail"></div>
        </div>
      </div>
    </div>

    <!-- Right: Jobs + Config -->
    <div style="display:flex;flex-direction:column;gap:16px">
      <div class="card">
        <div class="card-title" style="justify-content:space-between;display:flex;align-items:center">
          <span>Jobs</span>
          <button onclick="loadJobs()" style="background:none;border:none;color:var(--muted);cursor:pointer;font-size:0.75rem">↻ refresh</button>
        </div>
        <div class="jobs-list" id="jobsList">
          <div class="empty-state">No jobs yet. Upload a video to get started.</div>
        </div>
      </div>

      <div class="card">
        <div class="card-title">Server Config</div>
        <div class="config-panel" id="configPanel">
          <div style="color:var(--muted);font-size:0.8rem">Loading...</div>
        </div>
      </div>
    </div>
  </main>
</div>

<div id="toast"></div>

<script>
const API = '/api/v1';
let selectedFile = null;
let activeProfile = 'balanced';
let pollInterval = null;
let profiles = {};
let editorObjectURL = '';
let videoDuration = 0;
let cutRanges = [];
let markStart = null;
let markEnd = null;
let dragState = null;
let cutPreview = null;
let recorderStream = null;
let mediaRecorder = null;
let recordedChunks = [];
let recorderTimer = null;
let recorderStartedAt = 0;

// ── Init ──────────────────────────────────────────────────────────────────
document.addEventListener('DOMContentLoaded', async () => {
  await loadProfiles();
  await loadJobs();
  await loadConfig();
  checkHealth();
  setInterval(checkHealth, 15000);
  setInterval(loadJobs, 5000);
  setupDragDrop();
});

// ── Health ────────────────────────────────────────────────────────────────
async function checkHealth() {
  try {
    const r = await fetch(API + '/health');
    const d = await r.json();
    const dot = document.getElementById('healthDot');
    dot.style.background = d.status === 'ok' ? 'var(--accent3)' : 'var(--warn)';
    const s = d.queue;
    document.getElementById('statsLabel').textContent =
      'workers active · ' + s.running + ' running · ' + s.queued + ' queued';
  } catch {}
}

// ── Profiles ──────────────────────────────────────────────────────────────
async function loadProfiles() {
  try {
    const r = await fetch(API + '/profiles');
    profiles = await r.json();
    renderProfiles();
    applyProfile('balanced');
  } catch(e) { console.error(e); }
}

function renderProfiles() {
  const grid = document.getElementById('profilesGrid');
  const order = ['fast', 'balanced', 'quality', 'tiny', 'hd', 'custom'];
  const icons = { fast:'⚡', balanced:'⚖️', quality:'💎', tiny:'🪶', hd:'🎯', custom:'🔧' };
  const all = [...order.filter(k => profiles[k]), ...Object.keys(profiles).filter(k => !order.includes(k))];
  grid.innerHTML = all.map(name => {
    const p = profiles[name];
    return '<button class="profile-btn' + (name === activeProfile ? ' active' : '') + '" ' +
      'onclick="applyProfile(\'' + name + '\')" data-profile="' + name + '">' +
      (icons[name]||'📋') + ' ' + name.charAt(0).toUpperCase() + name.slice(1) +
      '<div class="profile-desc">' + (p.description||'') + '</div></button>';
  }).join('');
}

function applyProfile(name) {
  activeProfile = name;
  const p = profiles[name];
  if (!p) return;

  document.querySelectorAll('.profile-btn').forEach(b => b.classList.toggle('active', b.dataset.profile === name));

  setVal('fps', p.fps); syncSlider('fps','fpsVal');
  setVal('colors', p.colors); syncSlider('colors','colorsVal');
  setVal('width', p.width || 640);
  setVal('height', p.height || -1);
  setVal('dither', p.dither);
  setVal('bayerScale', p.bayer_scale || 2); syncSlider('bayerScale','bayerVal');
  setVal('startTime', p.start_time || '');
  setVal('duration', p.duration || '');
  setVal('speed', p.speed_multiplier || 1); syncSlider('speed','speedVal');
  setVal('loop', p.loop !== undefined ? p.loop : 0);
  document.getElementById('optimizePalette').checked = !!p.optimize_palette;
}

function setVal(id, v) {
  const el = document.getElementById(id);
  if (el) el.value = v;
}

// ── Sliders ───────────────────────────────────────────────────────────────
function syncSlider(sliderId, valId) {
  const v = parseFloat(document.getElementById(sliderId).value);
  const el = document.getElementById(valId);
  if (valId === 'speedVal') el.textContent = v + '×';
  else el.textContent = v;
}

// ── Screen Recorder ───────────────────────────────────────────────────────
function pickRecorderMimeType() {
  const candidates = [
    'video/webm;codecs=vp8,opus',
    'video/webm',
    'video/webm;codecs=vp9,opus',
  ];
  for (const type of candidates) {
    if (window.MediaRecorder && MediaRecorder.isTypeSupported(type)) {
      return type;
    }
  }
  return '';
}

async function startScreenRecording() {
  const getDisplayMedia = (navigator.mediaDevices && navigator.mediaDevices.getDisplayMedia)
    ? navigator.mediaDevices.getDisplayMedia.bind(navigator.mediaDevices)
    : (navigator.getDisplayMedia ? navigator.getDisplayMedia.bind(navigator) : null);

  if (!getDisplayMedia) {
    const reason = window.isSecureContext
      ? 'getDisplayMedia API is unavailable'
      : 'this page is not in a secure context (use https or localhost)';
    toast('Screen recording unavailable: ' + reason, 'error');
    return;
  }
  if (mediaRecorder && mediaRecorder.state === 'recording') return;
  try {
    recorderStream = await getDisplayMedia({
      video: true,
      audio: true,
    });
    const mimeType = pickRecorderMimeType();
    mediaRecorder = mimeType ? new MediaRecorder(recorderStream, { mimeType }) : new MediaRecorder(recorderStream);
    recordedChunks = [];

    mediaRecorder.ondataavailable = event => {
      if (event.data && event.data.size > 0) {
        recordedChunks.push(event.data);
      }
    };

    mediaRecorder.onstop = () => {
      stopRecorderTimer();
      setRecorderStatus('Recorder idle', false);
      toggleRecorderButtons(false);
      stopRecorderTracks();

      if (!recordedChunks.length) {
        toast('Recording is empty', 'error');
        return;
      }
      const type = mediaRecorder.mimeType || 'video/webm';
      const ext = type.includes('webm') ? 'webm' : 'mp4';
      const blob = new Blob(recordedChunks, { type });
      const file = new File([blob], 'screen-recording-' + Date.now() + '.' + ext, { type });
      handleFile(file);
      toast('Recording captured. Now cut ranges and save.', 'success');
    };

    mediaRecorder.onerror = e => {
      toast('Recorder error: ' + (e.error?.message || 'unknown'), 'error');
    };

    const vt = recorderStream.getVideoTracks()[0];
    if (vt) {
      vt.onended = () => {
        if (mediaRecorder && mediaRecorder.state === 'recording') {
          stopScreenRecording();
        }
      };
    }

    mediaRecorder.start(300);
    recorderStartedAt = Date.now();
    toggleRecorderButtons(true);
    startRecorderTimer();
    setRecorderStatus('REC 00:00', true);
    toast('Screen recording started', 'success');
  } catch (err) {
    setRecorderStatus('Recorder idle', false);
    toggleRecorderButtons(false);
    toast('Could not start recording: ' + err.message, 'error');
  }
}

function stopScreenRecording() {
  if (mediaRecorder && mediaRecorder.state === 'recording') {
    mediaRecorder.stop();
    setRecorderStatus('Finalizing recording...', false);
    return;
  }
  stopRecorderTracks();
  stopRecorderTimer();
  toggleRecorderButtons(false);
  setRecorderStatus('Recorder idle', false);
}

function stopRecorderTracks() {
  if (!recorderStream) return;
  recorderStream.getTracks().forEach(t => t.stop());
  recorderStream = null;
}

function toggleRecorderButtons(isRecording) {
  document.getElementById('recordStartBtn').disabled = isRecording;
  document.getElementById('recordStopBtn').disabled = !isRecording;
}

function setRecorderStatus(text, recording) {
  const el = document.getElementById('recorderStatus');
  el.textContent = text;
  el.classList.toggle('recording', !!recording);
}

function startRecorderTimer() {
  stopRecorderTimer();
  recorderTimer = setInterval(() => {
    const sec = Math.floor((Date.now() - recorderStartedAt) / 1000);
    setRecorderStatus('REC ' + formatClock(sec), true);
  }, 200);
}

function stopRecorderTimer() {
  if (!recorderTimer) return;
  clearInterval(recorderTimer);
  recorderTimer = null;
}

// ── File handling ─────────────────────────────────────────────────────────
function setupDragDrop() {
  const dz = document.getElementById('dropzone');
  const input = document.getElementById('fileInput');
  const scrub = document.getElementById('timelineScrub');
  const video = document.getElementById('editorVideo');
  const track = document.getElementById('timelineTrack');

  dz.addEventListener('dragover', e => { e.preventDefault(); dz.classList.add('dragover'); });
  dz.addEventListener('dragleave', () => dz.classList.remove('dragover'));
  dz.addEventListener('drop', e => {
    e.preventDefault();
    dz.classList.remove('dragover');
    const f = e.dataTransfer.files[0];
    if (f) handleFile(f);
  });
  input.addEventListener('change', () => { if (input.files[0]) handleFile(input.files[0]); });

  scrub.addEventListener('input', () => {
    if (!videoDuration) return;
    const t = parseFloat(scrub.value) || 0;
    if (video) video.currentTime = t;
    updateEditorHUD(t);
  });

  video.addEventListener('loadedmetadata', () => {
    setEditorDuration(Number.isFinite(video.duration) ? video.duration : 0);
    scrub.value = '0';
    cutRanges = [];
    markStart = null;
    markEnd = null;
    renderTimeline();
    renderSegments();
    updateEditorHUD(0);
    ensurePlayableDuration(video);
  });

  video.addEventListener('durationchange', () => {
    if (!videoDuration && Number.isFinite(video.duration) && video.duration > 0) {
      setEditorDuration(video.duration);
      renderTimeline();
      renderSegments();
      updateEditorHUD(video.currentTime || 0);
    }
  });

  video.addEventListener('error', () => {
    toast('Video preview failed to load. You can still trim if duration probe succeeds.', 'error');
  });

  video.addEventListener('timeupdate', () => {
    if (!videoDuration) return;
    if (cutPreview && video.currentTime >= cutPreview.end) {
      video.pause();
      stopCutPreview();
    }
    scrub.value = video.currentTime.toFixed(2);
    updateEditorHUD(video.currentTime);
  });

  document.addEventListener('mousemove', handleCutDrag);
  document.addEventListener('mouseup', endCutDrag);
  track.addEventListener('mouseleave', endCutDrag);
}

function handleFile(file) {
  selectedFile = file;
  document.getElementById('previewName').textContent = file.name;
  document.getElementById('previewMeta').textContent =
    formatBytes(file.size) + ' · ' + (file.type || 'video');
  document.getElementById('filePreview').style.display = 'flex';
  document.getElementById('convertBtn').disabled = false;
  document.getElementById('saveBtn').disabled = false;
  initEditor(file);
}

function clearFile() {
  stopCutPreview();
  selectedFile = null;
  document.getElementById('fileInput').value = '';
  document.getElementById('filePreview').style.display = 'none';
  document.getElementById('convertBtn').disabled = true;
  document.getElementById('saveBtn').disabled = true;
  document.getElementById('progressWrap').style.display = 'none';
  document.getElementById('editorWrap').style.display = 'none';
  document.getElementById('segmentList').innerHTML = '';
  cutRanges = [];
  markStart = null;
  markEnd = null;
  videoDuration = 0;
  if (editorObjectURL) {
    URL.revokeObjectURL(editorObjectURL);
    editorObjectURL = '';
  }
}

function initEditor(file) {
  const wrap = document.getElementById('editorWrap');
  const video = document.getElementById('editorVideo');
  if (editorObjectURL) URL.revokeObjectURL(editorObjectURL);
  editorObjectURL = URL.createObjectURL(file);
  video.src = editorObjectURL;
  wrap.style.display = 'block';
  setEditorDuration(0);
  probeVideoDuration(file);
}

function setEditorDuration(duration) {
  const scrub = document.getElementById('timelineScrub');
  videoDuration = (Number.isFinite(duration) && duration > 0) ? duration : 0;
  scrub.max = videoDuration ? videoDuration.toFixed(2) : '0';
  document.getElementById('cutStart').value = '0';
  document.getElementById('cutEnd').value = videoDuration ? videoDuration.toFixed(2) : '0';
}

async function probeVideoDuration(file) {
  try {
    const form = new FormData();
    form.append('video', file, file.name || 'recording.webm');
    const r = await fetch(API + '/probe', { method: 'POST', body: form });
    if (!r.ok) return;
    const info = await r.json();
    if (!videoDuration && info && Number.isFinite(info.duration) && info.duration > 0) {
      setEditorDuration(info.duration);
      renderTimeline();
      renderSegments();
      updateEditorHUD(0);
    }
  } catch {}
}

function ensurePlayableDuration(video) {
  if (!video) return;
  if (Number.isFinite(video.duration) && video.duration > 0) return;
  const onSeeked = () => {
    if (Number.isFinite(video.duration) && video.duration > 0) {
      if (!videoDuration) {
        setEditorDuration(video.duration);
        renderTimeline();
        renderSegments();
      }
      video.currentTime = 0;
    }
    video.removeEventListener('seeked', onSeeked);
  };
  try {
    video.addEventListener('seeked', onSeeked);
    video.currentTime = 1e9;
  } catch {
    video.removeEventListener('seeked', onSeeked);
  }
}

function markCutStart() {
  const video = document.getElementById('editorVideo');
  markStart = (video && Number.isFinite(video.currentTime)) ? video.currentTime : 0;
  document.getElementById('cutStart').value = markStart.toFixed(2);
}

function markCutEnd() {
  const video = document.getElementById('editorVideo');
  markEnd = (video && Number.isFinite(video.currentTime)) ? video.currentTime : 0;
  document.getElementById('cutEnd').value = markEnd.toFixed(2);
}

function cutMarkedRange() {
  if (markStart === null || markEnd === null) {
    toast('Mark both cut start and cut end first', 'error');
    return;
  }
  addCutRange(markStart, markEnd);
}

function addCutFromInputs() {
  const s = parseFloat(document.getElementById('cutStart').value);
  const e = parseFloat(document.getElementById('cutEnd').value);
  addCutRange(s, e);
}

function addCutRange(start, end) {
  stopCutPreview();
  if (!videoDuration) {
    toast('Video duration is not ready yet. Wait for preview/probe to finish.', 'error');
    return;
  }
  if (!Number.isFinite(start) || !Number.isFinite(end)) {
    toast('Invalid cut range', 'error');
    return;
  }
  let s = Math.max(0, Math.min(start, end));
  let e = Math.min(videoDuration, Math.max(start, end));
  if (e - s < 0.05) {
    toast('Cut range too small', 'error');
    return;
  }

  cutRanges.push({ start: s, end: e });
  cutRanges = mergeRanges(cutRanges);
  syncCutBounds();
  document.getElementById('cutStart').value = s.toFixed(2);
  document.getElementById('cutEnd').value = e.toFixed(2);
  renderTimeline();
  renderSegments();
}

function resetCuts() {
  stopCutPreview();
  cutRanges = [];
  renderTimeline();
  renderSegments();
}

function removeCut(index) {
  if (index < 0 || index >= cutRanges.length) return;
  stopCutPreview();
  cutRanges.splice(index, 1);
  cutRanges = mergeRanges(cutRanges);
  renderTimeline();
  renderSegments();
}

function applyCutRow(index) {
  if (index < 0 || index >= cutRanges.length) return;
  const s = parseFloat(document.getElementById('cutRowStart_' + index).value);
  const e = parseFloat(document.getElementById('cutRowEnd_' + index).value);
  if (!Number.isFinite(s) || !Number.isFinite(e)) {
    toast('Invalid cut range values', 'error');
    return;
  }
  const start = Math.max(0, Math.min(s, e));
  const end = Math.min(videoDuration, Math.max(s, e));
  if (end-start < 0.05) {
    toast('Cut range too small', 'error');
    return;
  }
  cutRanges[index] = { start, end };
  cutRanges = mergeRanges(cutRanges);
  syncCutBounds();
  renderTimeline();
  renderSegments();
}

function resetCutRow(index) {
  if (index < 0 || index >= cutRanges.length) return;
  document.getElementById('cutRowStart_' + index).value = cutRanges[index].start.toFixed(2);
  document.getElementById('cutRowEnd_' + index).value = cutRanges[index].end.toFixed(2);
}

function playCut(index) {
  if (index < 0 || index >= cutRanges.length) return;
  const video = document.getElementById('editorVideo');
  if (!video) return;
  stopCutPreview();
  const cut = cutRanges[index];
  cutPreview = { index, end: cut.end };
  video.currentTime = cut.start;
  video.play().catch(() => {});
}

function stopCutPreview() {
  cutPreview = null;
}

function mergeRanges(ranges) {
  const sorted = [...ranges]
    .filter(r => Number.isFinite(r.start) && Number.isFinite(r.end) && r.end > r.start)
    .sort((a, b) => a.start - b.start);
  if (!sorted.length) return [];
  const merged = [sorted[0]];
  for (let i = 1; i < sorted.length; i++) {
    const last = merged[merged.length - 1];
    const cur = sorted[i];
    if (cur.start <= last.end) {
      last.end = Math.max(last.end, cur.end);
    } else {
      merged.push(cur);
    }
  }
  return merged;
}

function getKeepSegments() {
  if (!videoDuration) return [];
  const cuts = mergeRanges(cutRanges);
  if (!cuts.length) return [{ start: 0, end: videoDuration }];

  const keep = [];
  let cursor = 0;
  for (const cut of cuts) {
    if (cut.start > cursor) keep.push({ start: cursor, end: cut.start });
    cursor = Math.max(cursor, cut.end);
  }
  if (cursor < videoDuration) keep.push({ start: cursor, end: videoDuration });
  return keep.filter(s => (s.end - s.start) >= 0.05).map(s => ({
    start: Number(s.start.toFixed(3)),
    end: Number(s.end.toFixed(3)),
  }));
}

function renderTimeline() {
  const track = document.getElementById('timelineTrack');
  const playhead = document.getElementById('timelinePlayhead');
  if (!track || !playhead || !videoDuration) return;

  track.querySelectorAll('.timeline-seg,.timeline-cut').forEach(n => n.remove());
  const keeps = getKeepSegments();
  keeps.forEach(seg => {
    const node = document.createElement('div');
    node.className = 'timeline-seg';
    node.style.left = ((seg.start / videoDuration) * 100) + '%';
    node.style.width = (((seg.end - seg.start) / videoDuration) * 100) + '%';
    track.appendChild(node);
  });

  cutRanges.forEach((cut, idx) => {
    const node = document.createElement('div');
    node.className = 'timeline-cut';
    node.style.left = ((cut.start / videoDuration) * 100) + '%';
    node.style.width = (((cut.end - cut.start) / videoDuration) * 100) + '%';

    const left = document.createElement('div');
    left.className = 'timeline-handle left';
    left.addEventListener('mousedown', ev => startCutDrag(ev, idx, 'left'));

    const right = document.createElement('div');
    right.className = 'timeline-handle right';
    right.addEventListener('mousedown', ev => startCutDrag(ev, idx, 'right'));

    const label = document.createElement('div');
    label.className = 'timeline-cut-label';
    label.textContent = 'Cut ' + (idx + 1);

    node.appendChild(left);
    node.appendChild(right);
    if ((cut.end - cut.start) > 0.35) {
      node.appendChild(label);
    }
    track.appendChild(node);
  });
  updatePlayhead();
}

function renderSegments() {
  const list = document.getElementById('segmentList');
  if (!videoDuration) {
    list.innerHTML = '';
    return;
  }
  if (!cutRanges.length) {
    list.innerHTML = '<span class="segment-empty">No cuts yet. Add as many cut ranges as you need.</span>';
    return;
  }

  list.innerHTML = cutRanges.map((s, idx) =>
    '<div class="segment-row">' +
      '<span class="segment-row-id">#' + (idx + 1) + '</span>' +
      '<input class="segment-input" id="cutRowStart_' + idx + '" type="number" min="0" step="0.01" value="' + s.start.toFixed(2) + '">' +
      '<input class="segment-input" id="cutRowEnd_' + idx + '" type="number" min="0" step="0.01" value="' + s.end.toFixed(2) + '">' +
      '<button class="chip-btn warn" type="button" onclick="playCut(' + idx + ')">Play Cut</button>' +
      '<button class="chip-btn" type="button" onclick="applyCutRow(' + idx + ')">Update</button>' +
      '<button class="chip-btn" type="button" onclick="resetCutRow(' + idx + ')">Reset</button>' +
      '<button class="chip-btn danger" type="button" onclick="removeCut(' + idx + ')">Remove</button>' +
    '</div>'
  ).join('');
}

function startCutDrag(event, index, edge) {
  if (!videoDuration || index < 0 || index >= cutRanges.length) return;
  event.preventDefault();
  stopCutPreview();
  dragState = {
    index,
    edge,
    startX: event.clientX,
    start: cutRanges[index].start,
    end: cutRanges[index].end,
  };
}

function handleCutDrag(event) {
  if (!dragState || !videoDuration) return;
  const track = document.getElementById('timelineTrack');
  if (!track) return;
  const width = track.clientWidth || 1;
  const deltaSec = ((event.clientX - dragState.startX) / width) * videoDuration;
  const idx = dragState.index;
  const prevEnd = idx > 0 ? cutRanges[idx - 1].end : 0;
  const nextStart = idx < cutRanges.length - 1 ? cutRanges[idx + 1].start : videoDuration;
  const minDur = 0.05;

  if (dragState.edge === 'left') {
    const maxStart = Math.min(dragState.end - minDur, nextStart - minDur);
    cutRanges[idx].start = clamp(dragState.start + deltaSec, prevEnd, maxStart);
  } else {
    const minEnd = Math.max(dragState.start + minDur, prevEnd + minDur);
    cutRanges[idx].end = clamp(dragState.end + deltaSec, minEnd, nextStart);
  }
  syncCutBounds();
  renderTimeline();
  renderSegments();
}

function endCutDrag() {
  if (!dragState) return;
  dragState = null;
  cutRanges = mergeRanges(cutRanges);
  syncCutBounds();
  renderTimeline();
  renderSegments();
}

function syncCutBounds() {
  for (let i = 0; i < cutRanges.length; i++) {
    cutRanges[i].start = Number(clamp(cutRanges[i].start, 0, videoDuration).toFixed(3));
    cutRanges[i].end = Number(clamp(cutRanges[i].end, 0, videoDuration).toFixed(3));
  }
}

function updatePlayhead() {
  const video = document.getElementById('editorVideo');
  const playhead = document.getElementById('timelinePlayhead');
  if (!video || !playhead || !videoDuration) return;
  const pct = Math.max(0, Math.min(100, (video.currentTime / videoDuration) * 100));
  playhead.style.left = 'calc(' + pct + '% - 1px)';
}

function updateEditorHUD(time) {
  document.getElementById('editorNow').textContent = formatTime(time) + ' / ' + formatTime(videoDuration);
  updatePlayhead();
}

// ── Conversion ────────────────────────────────────────────────────────────
async function saveEditedVideo() {
  if (!selectedFile) return;
  const btn = document.getElementById('saveBtn');
  btn.disabled = true;
  btn.textContent = 'Saving...';

  const cutPayload = cutRanges.map(s => ({
    start: Number(s.start.toFixed(3)),
    end: Number(s.end.toFixed(3)),
  }));

  const form = new FormData();
  form.append('video', selectedFile, selectedFile.name || 'recording.webm');
  form.append('cut_ranges', JSON.stringify(cutPayload));
  if (videoDuration > 0) {
    form.append('duration_hint', String(Number(videoDuration.toFixed(3))));
  }

  try {
    const r = await fetch(API + '/save-edited', { method: 'POST', body: form });
    if (!r.ok) {
      const err = await r.json().catch(() => ({}));
      throw new Error(err.error || ('HTTP ' + r.status));
    }
    const blob = await r.blob();
    const cd = r.headers.get('content-disposition') || '';
    const filename = parseDownloadFilename(cd) || 'edited-video.mp4';
    downloadBlob(blob, filename);
    toast('Edited video saved (cut ranges removed)', 'success');
  } catch (e) {
    toast('Save failed: ' + e.message, 'error');
  } finally {
    btn.disabled = false;
    btn.textContent = 'Save Edited Video';
  }
}

async function startConvert() {
  if (!selectedFile) return;

  const btn = document.getElementById('convertBtn');
  const saveBtn = document.getElementById('saveBtn');
  btn.disabled = true;
  saveBtn.disabled = true;
  btn.textContent = 'Uploading...';

  const params = {
    fps: parseFloat(document.getElementById('fps').value),
    width: parseInt(document.getElementById('width').value),
    height: parseInt(document.getElementById('height').value),
    colors: parseInt(document.getElementById('colors').value),
    dither: document.getElementById('dither').value,
    bayer_scale: parseInt(document.getElementById('bayerScale').value),
    keep_segments: getKeepSegments(),
    start_time: '',
    duration: '',
    speed_multiplier: parseFloat(document.getElementById('speed').value),
    loop: parseInt(document.getElementById('loop').value),
    optimize_palette: document.getElementById('optimizePalette').checked,
    stats_mode: 'diff',
    name: activeProfile,
  };

  const form = new FormData();
  form.append('video', selectedFile);
  form.append('params', JSON.stringify(params));

  try {
    const r = await fetch(API + '/convert', { method: 'POST', body: form });
    const job = await r.json();

    if (!r.ok) {
      toast('Upload failed: ' + (job.error || r.status), 'error');
      btn.disabled = false;
      saveBtn.disabled = false;
      btn.textContent = 'Convert to GIF';
      return;
    }

    btn.textContent = 'Processing...';
    showProgress(true);
    pollJob(job.id);
    toast('Job submitted! ID: ' + job.id.slice(0,8) + '…', 'success');
  } catch(e) {
    toast('Network error: ' + e.message, 'error');
    btn.disabled = false;
    saveBtn.disabled = false;
    btn.textContent = 'Convert to GIF';
  }
}

function pollJob(id) {
  clearInterval(pollInterval);
  let dots = 0;
  pollInterval = setInterval(async () => {
    try {
      const r = await fetch(API + '/jobs/' + id);
      const job = await r.json();
      updateProgress(job, dots++);

      if (job.status === 'done') {
        clearInterval(pollInterval);
        setProgress(100, 'Complete!', '✅ GIF ready for download');
        toast('🎉 Conversion complete!', 'success');
        document.getElementById('convertBtn').textContent = 'Convert to GIF';
        document.getElementById('convertBtn').disabled = false;
        document.getElementById('saveBtn').disabled = false;
        loadJobs();
        setTimeout(() => triggerDownload(id, job), 500);
      } else if (job.status === 'failed') {
        clearInterval(pollInterval);
        setProgress(0, 'Failed', '❌ ' + (job.error || 'Unknown error'));
        toast('Conversion failed: ' + (job.error||''), 'error');
        document.getElementById('convertBtn').textContent = 'Convert to GIF';
        document.getElementById('convertBtn').disabled = false;
        document.getElementById('saveBtn').disabled = false;
        loadJobs();
      }
    } catch {}
  }, 1200);
}

function updateProgress(job, tick) {
  const statusLabels = { queued:'Queued...', running:'Processing with FFmpeg...', done:'Complete!', failed:'Failed' };
  const pct = job.status === 'running' ? Math.min(90, tick * 8) : job.status === 'done' ? 100 : 0;
  const detail = job.status === 'running'
    ? ['Generating palette...','Rendering frames...','Applying dithering...','Optimizing GIF...'][tick % 4]
    : '';
  setProgress(pct, statusLabels[job.status] || job.status, detail);
}

function setProgress(pct, label, detail) {
  document.getElementById('progressFill').style.width = pct + '%';
  document.getElementById('progressStatus').textContent = label;
  document.getElementById('progressPct').textContent = pct ? pct + '%' : '—';
  document.getElementById('progressDetail').textContent = detail;
}

function showProgress(show) {
  document.getElementById('progressWrap').style.display = show ? 'block' : 'none';
  if (show) setProgress(0, 'Queued...', 'Waiting for worker...');
}

// ── Jobs list ─────────────────────────────────────────────────────────────
async function loadJobs() {
  try {
    const r = await fetch(API + '/jobs');
    const jobs = await r.json();
    renderJobs(jobs);
  } catch {}
}

function renderJobs(jobs) {
  const el = document.getElementById('jobsList');
  if (!jobs || !jobs.length) {
    el.innerHTML = '<div class="empty-state">No jobs yet.</div>';
    return;
  }
  jobs.sort((a,b) => new Date(b.created_at) - new Date(a.created_at));
  el.innerHTML = jobs.slice(0,20).map(j => {
    const size = j.result ? formatBytes(j.result.output_size) : '';
    const dur  = j.result ? j.result.duration : '';
    const meta = [j.profile.name, size, dur].filter(Boolean).join(' · ');
    return '<div class="job-item">' +
      '<div class="job-status-dot status-' + j.status + '"></div>' +
      '<div class="job-info">' +
        '<div class="job-name">' + escHtml(j.file_name || 'Unknown') + '</div>' +
        '<div class="job-meta">' + j.status.toUpperCase() + (meta ? ' · ' + meta : '') + '</div>' +
      '</div>' +
      '<div class="job-actions">' +
        (j.status === 'done'
          ? '<button class="btn-sm btn-download" onclick="triggerDownload(\'' + j.id + '\',null)">⬇ GIF</button>'
          : '') +
        '<button class="btn-sm btn-delete" onclick="deleteJob(\'' + j.id + '\')">✕</button>' +
      '</div></div>';
  }).join('');
}

async function deleteJob(id) {
  await fetch(API + '/jobs/' + id, { method: 'DELETE' });
  loadJobs();
}

function triggerDownload(id) {
  const a = document.createElement('a');
  a.href = API + '/jobs/' + id + '/download';
  a.click();
}

// ── Config display ────────────────────────────────────────────────────────
async function loadConfig() {
  try {
    const r = await fetch(API + '/config');
    const cfg = await r.json();
    const panel = document.getElementById('configPanel');
    const rows = [
      ['Workers', cfg.queue?.workers],
      ['Max Upload', formatBytes(cfg.server?.max_upload_bytes)],
      ['Upload Dir', cfg.storage?.upload_dir],
      ['Output Dir', cfg.storage?.output_dir],
      ['Job Timeout', cfg.queue?.job_timeout_sec + 's'],
      ['File TTL', cfg.storage?.max_age_hours + 'h'],
    ];
    panel.innerHTML = rows.map(([k,v]) =>
      '<div class="config-row"><span class="config-key">' + k + '</span><span class="config-val">' + v + '</span></div>'
    ).join('');
  } catch {}
}

// ── Utils ─────────────────────────────────────────────────────────────────
function formatBytes(b) {
  if (!b) return '0 B';
  const u = ['B','KB','MB','GB'];
  let i = 0;
  while (b >= 1024 && i < u.length - 1) { b /= 1024; i++; }
  return b.toFixed(1) + ' ' + u[i];
}

function clamp(v, min, max) {
  if (v < min) return min;
  if (v > max) return max;
  return v;
}

function formatTime(sec) {
  if (!Number.isFinite(sec) || sec < 0) sec = 0;
  const m = Math.floor(sec / 60);
  const s = sec - (m * 60);
  return String(m).padStart(2, '0') + ':' + s.toFixed(2).padStart(5, '0');
}

function formatClock(totalSec) {
  const m = Math.floor(totalSec / 60);
  const s = totalSec % 60;
  return String(m).padStart(2, '0') + ':' + String(s).padStart(2, '0');
}

function downloadBlob(blob, filename) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}

function parseDownloadFilename(contentDisposition) {
  const m = /filename=\"?([^\";]+)\"?/i.exec(contentDisposition || '');
  return m ? m[1] : '';
}

function escHtml(s) {
  return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

function toast(msg, type='success') {
  const el = document.getElementById('toast');
  el.textContent = msg;
  el.className = 'show ' + type;
  setTimeout(() => el.classList.remove('show'), 3500);
}
</script>
</body>
</html>`
