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

  .container { margin: 0 auto; padding: 0 24px; position: relative; z-index: 1; }

  header {
    padding: 32px 0 24px;
    border-bottom: 1px solid var(--border);
  }

  .tabs {
    display: flex;
    gap: 10px;
    margin-top: 16px;
  }

  .workflow-picker {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 18px;
    padding: 28px 0 10px;
  }

  @media (max-width: 900px) {
    .workflow-picker {
      grid-template-columns: 1fr;
    }
  }

  .workflow-card {
    position: relative;
    overflow: hidden;
    border: 1px solid var(--border);
    border-radius: 22px;
    background: linear-gradient(180deg, rgba(17,17,24,0.98), rgba(10,12,20,0.98));
    padding: 26px;
    cursor: pointer;
    text-align: left;
    color: var(--text);
    transition: transform 0.2s, border-color 0.2s, box-shadow 0.2s;
  }

  .workflow-card::before {
    content: '';
    position: absolute;
    inset: 0;
    pointer-events: none;
    opacity: 0.9;
  }

  .workflow-card.screenshare::before {
    background:
      radial-gradient(circle at top right, rgba(52,211,153,0.18), transparent 40%),
      linear-gradient(135deg, rgba(16,185,129,0.16), transparent 55%);
  }

  .workflow-card.video2gif::before {
    background:
      radial-gradient(circle at top right, rgba(167,139,250,0.2), transparent 40%),
      linear-gradient(135deg, rgba(124,58,237,0.16), transparent 55%);
  }

  .workflow-card:hover {
    transform: translateY(-3px);
    border-color: rgba(167,139,250,0.65);
    box-shadow: 0 18px 42px rgba(0,0,0,0.28);
  }

  .workflow-card > * {
    position: relative;
    z-index: 1;
  }

  .workflow-eyebrow {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    color: var(--muted);
    font-size: 0.72rem;
    text-transform: uppercase;
    letter-spacing: 0.14em;
    font-family: 'JetBrains Mono', monospace;
  }

  .workflow-title {
    margin-top: 14px;
    font-size: 1.65rem;
    font-weight: 700;
    letter-spacing: -0.03em;
  }

  .workflow-copy {
    margin-top: 12px;
    color: #cbd5e1;
    font-size: 0.92rem;
    line-height: 1.65;
    max-width: 42ch;
  }

  .workflow-features {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
    margin-top: 20px;
  }

  @media (max-width: 640px) {
    .workflow-features {
      grid-template-columns: 1fr;
    }
  }

  .workflow-feature {
    border: 1px solid rgba(148,163,184,0.18);
    border-radius: 12px;
    padding: 12px;
    background: rgba(15,20,32,0.72);
  }

  .workflow-feature strong {
    display: block;
    font-size: 0.82rem;
    margin-bottom: 5px;
  }

  .workflow-feature span {
    display: block;
    color: var(--muted);
    font-size: 0.76rem;
    line-height: 1.55;
  }

  .workflow-cta {
    margin-top: 22px;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.78rem;
    color: var(--text);
  }

  .tab-btn {
    border: 1px solid var(--border);
    background: transparent;
    color: var(--muted);
    border-radius: 8px;
    padding: 8px 14px;
    cursor: pointer;
    font-size: 0.82rem;
    font-family: 'JetBrains Mono', monospace;
  }

  .tab-btn.active {
    border-color: var(--accent);
    background: rgba(124,58,237,0.16);
    color: var(--accent2);
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

  main { padding: 20px 0 32px; display: grid; grid-template-columns: 340px 1fr; gap: 18px; align-items: start; }

  main.setup-only {
    grid-template-columns: minmax(0, 760px);
    justify-content: center;
  }

  main.setup-only #sidebarCol {
    position: static;
    top: auto;
  }

  main.setup-only #mainCol {
    display: none;
  }

  @media (max-width: 800px) { main { grid-template-columns: 1fr; } }

  #sidebarCol {
    position: sticky;
    top: 12px;
  }

  #mainCol {
    display: flex;
    flex-direction: column;
    gap: 14px;
    min-width: 0;
  }

  .editor-card {
    padding: 16px;
  }

  .shared-mode main {
    grid-template-columns: 1fr;
  }

  .shared-mode #sidebarCol,
  .shared-mode .tabs,
  .shared-mode #rightPanel {
    display: none !important;
  }

  .shared-mode .container {
    max-width: 1680px;
  }

  .shared-mode .editor-video {
    max-height: 72vh;
    min-height: 460px;
  }

  .shared-mode .editor-toolbar,
  .shared-mode .timeline-wrap,
  .shared-mode .cut-controls,
  .shared-mode .segment-title,
  .shared-mode .segment-list,
  .shared-mode .comment-tools,
  .shared-mode .comment-actions {
    display: none !important;
  }

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
    flex-wrap: wrap;
  }

  .recorder-status {
    margin-left: auto;
    font-size: 0.75rem;
    color: var(--muted);
    font-family: 'JetBrains Mono', monospace;
    display: inline-flex;
    align-items: center;
    gap: 7px;
    padding: 6px 10px;
    border: 1px solid var(--border);
    border-radius: 999px;
    background: rgba(15,20,32,0.75);
  }

  .recorder-status.recording {
    color: #fecaca;
    border-color: rgba(239,68,68,0.7);
    background: rgba(239,68,68,0.16);
  }

  .rec-indicator {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: #64748b;
    box-shadow: 0 0 0 0 rgba(100,116,139,0.6);
  }

  .recorder-status.recording .rec-indicator {
    background: #ef4444;
    box-shadow: 0 0 0 0 rgba(239,68,68,0.85);
    animation: recPulse 1.1s infinite;
  }

  @keyframes recPulse {
    0% { box-shadow: 0 0 0 0 rgba(239,68,68,0.7); }
    70% { box-shadow: 0 0 0 10px rgba(239,68,68,0); }
    100% { box-shadow: 0 0 0 0 rgba(239,68,68,0); }
  }

  .recorder-settings {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
    margin-bottom: 12px;
    padding: 10px;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: rgba(15,20,32,0.92);
  }

  .inline-check {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-size: 0.76rem;
    color: var(--text);
    font-family: 'JetBrains Mono', monospace;
  }

  .inline-check input {
    width: 15px;
    height: 15px;
    accent-color: var(--accent);
  }

  .draft-panel {
    margin-top: 12px;
    padding: 10px;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: rgba(15,20,32,0.92);
  }

  .draft-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-bottom: 8px;
  }

  .draft-copy {
    font-size: 0.74rem;
    color: var(--muted);
    line-height: 1.45;
  }

  .draft-actions {
    display: flex;
    gap: 8px;
    margin-top: 10px;
    flex-wrap: wrap;
  }

  .mode-copy {
    color: var(--muted);
    font-size: 0.8rem;
    line-height: 1.6;
    margin-bottom: 14px;
  }

  .save-video-wide {
    width: 100%;
    margin: 8px 0 12px;
    padding: 11px 12px;
    border-radius: 8px;
    border: 1px solid rgba(16,185,129,0.7);
    background: linear-gradient(90deg, rgba(16,185,129,0.22), rgba(5,150,105,0.22));
    color: #6ee7b7;
    font-family: 'JetBrains Mono', monospace;
    font-weight: 700;
    cursor: pointer;
  }

  .save-video-wide:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  .save-video-wide:hover:not(:disabled) {
    border-color: rgba(16,185,129,0.95);
    background: linear-gradient(90deg, rgba(16,185,129,0.3), rgba(5,150,105,0.3));
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
    max-height: 520px;
    min-height: 320px;
    border-radius: 8px;
    background: #000;
    object-fit: contain;
  }

  .video-stage {
    position: relative;
  }

  .video-overlay {
    position: absolute;
    inset: 0;
    z-index: 2;
    border-radius: 8px;
    cursor: crosshair;
    pointer-events: none;
  }

  .video-overlay.capture {
    pointer-events: auto;
  }

  .comment-dot {
    position: absolute;
    width: 10px;
    height: 10px;
    border-radius: 999px;
    background: #f59e0b;
    border: 1px solid #fff;
    transform: translate(-50%, -50%);
    box-shadow: 0 0 8px rgba(245,158,11,0.8);
  }

  .comment-dot.active {
    background: #34d399;
    box-shadow: 0 0 12px rgba(52,211,153,0.85);
  }

  .comment-dot.resolved {
    background: #60a5fa;
    box-shadow: 0 0 10px rgba(96,165,250,0.8);
  }

  .comment-layout {
    margin-top: 10px;
    display: grid;
    grid-template-columns: 1fr 280px;
    gap: 12px;
  }

  @media (max-width: 900px) {
    .comment-layout {
      grid-template-columns: 1fr;
    }
  }

  .comment-sidebar {
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 10px;
    background: #0e1320;
    min-height: 120px;
  }

  .comment-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
    max-height: 210px;
    overflow: auto;
    margin-top: 8px;
  }

  .comment-item {
    border: 1px solid var(--border);
    background: rgba(245,158,11,0.08);
    color: #fcd34d;
    border-radius: 6px;
    padding: 8px;
    font-size: 0.72rem;
    font-family: 'JetBrains Mono', monospace;
    cursor: pointer;
  }

  .comment-item.active {
    border-color: rgba(52,211,153,0.8);
    background: rgba(52,211,153,0.1);
    color: #6ee7b7;
  }

  .comment-item.resolved {
    background: rgba(96,165,250,0.08);
    color: #bfdbfe;
  }

  .comment-row {
    display: flex;
    align-items: center;
    gap: 8px;
    justify-content: space-between;
  }

  .comment-main {
    min-width: 0;
    flex: 1;
  }

  .comment-text {
    margin-top: 6px;
    line-height: 1.45;
    word-break: break-word;
  }

  .comment-actions {
    display: flex;
    gap: 6px;
    margin-top: 8px;
    flex-wrap: wrap;
  }

  .comment-mini-btn {
    border: 1px solid var(--border);
    background: transparent;
    color: inherit;
    border-radius: 999px;
    padding: 2px 8px;
    font-size: 0.64rem;
    cursor: pointer;
    font-family: 'JetBrains Mono', monospace;
  }

  .comment-badge {
    display: inline-flex;
    align-items: center;
    border-radius: 999px;
    padding: 2px 8px;
    font-size: 0.62rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    border: 1px solid rgba(245,158,11,0.45);
    color: #fbbf24;
  }

  .comment-badge.resolved {
    border-color: rgba(96,165,250,0.5);
    color: #93c5fd;
  }

  .comment-summary {
    margin-top: 8px;
    color: var(--muted);
    font-size: 0.7rem;
    font-family: 'JetBrains Mono', monospace;
  }

  .comment-tools {
    margin-top: 10px;
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 8px;
    align-items: center;
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

  .editor-summary {
    margin-top: 12px;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .summary-pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 10px;
    border-radius: 999px;
    background: rgba(15,20,32,0.92);
    border: 1px solid var(--border);
    font-size: 0.68rem;
    color: var(--muted);
    font-family: 'JetBrains Mono', monospace;
  }

  .shortcut-note {
    margin-top: 10px;
    color: var(--muted);
    font-size: 0.69rem;
    font-family: 'JetBrains Mono', monospace;
    line-height: 1.6;
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
    grid-template-columns: 42px 1fr 1fr auto auto auto auto auto;
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

  .bottom-save-wrap {
    margin-top: 18px;
    display: flex;
    justify-content: flex-end;
  }

  .bottom-save-wrap .save-video-wide {
    width: min(360px, 100%);
    margin: 0;
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

  .share-panel {
    margin-top: 14px;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: var(--surface2);
    padding: 10px;
  }

  .share-line {
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 8px;
    align-items: center;
  }

  .share-label {
    font-size: 0.72rem;
    color: var(--muted);
    text-transform: uppercase;
    letter-spacing: 0.08em;
    margin-bottom: 8px;
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

  .progress-steps {
    margin-top: 12px;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(108px, 1fr));
    gap: 8px;
  }

  .progress-step {
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 10px;
    background: rgba(15,20,32,0.92);
    color: var(--muted);
    font-size: 0.7rem;
    font-family: 'JetBrains Mono', monospace;
  }

  .progress-step.active {
    border-color: rgba(167,139,250,0.8);
    color: var(--accent2);
    box-shadow: inset 0 0 0 1px rgba(124,58,237,0.25);
  }

  .progress-step.done {
    border-color: rgba(16,185,129,0.75);
    color: #6ee7b7;
  }

  .progress-step.failed {
    border-color: rgba(239,68,68,0.8);
    color: #fca5a5;
  }

  .result-panel {
    display: none;
    margin-top: 16px;
    border: 1px solid rgba(52,211,153,0.3);
    border-radius: 12px;
    padding: 14px;
    background: linear-gradient(180deg, rgba(10,22,20,0.96), rgba(10,15,24,0.96));
  }

  .result-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .result-title {
    font-size: 0.88rem;
    font-weight: 700;
    color: #d1fae5;
  }

  .result-meta {
    font-size: 0.72rem;
    color: var(--muted);
    font-family: 'JetBrains Mono', monospace;
    margin-top: 4px;
  }

  .result-preview {
    margin-top: 12px;
  }

  .result-preview video,
  .result-preview img {
    width: 100%;
    max-height: 280px;
    border-radius: 10px;
    border: 1px solid var(--border);
    background: #000;
  }

  .result-actions {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
    margin-top: 12px;
  }

  .result-link {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    border-radius: 999px;
    border: 1px solid rgba(167,139,250,0.5);
    color: var(--text);
    text-decoration: none;
    font-size: 0.76rem;
    font-family: 'JetBrains Mono', monospace;
    background: rgba(124,58,237,0.14);
  }

  .result-link.download {
    border-color: rgba(16,185,129,0.6);
    background: rgba(16,185,129,0.14);
    color: #d1fae5;
  }

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

  .auth-gate {
    position: fixed;
    inset: 0;
    z-index: 9998;
    display: none;
    align-items: center;
    justify-content: center;
    padding: 24px;
    background: rgba(3,7,18,0.84);
    backdrop-filter: blur(10px);
  }

  .auth-gate.visible {
    display: flex;
  }

  .auth-card {
    width: min(420px, 100%);
    border: 1px solid var(--border);
    border-radius: 18px;
    background: linear-gradient(180deg, rgba(17,17,24,0.98), rgba(9,12,20,0.98));
    box-shadow: 0 30px 80px rgba(0,0,0,0.45);
    padding: 24px;
  }

  .auth-copy {
    color: var(--muted);
    font-size: 0.82rem;
    line-height: 1.55;
    margin: 10px 0 16px;
  }

  .auth-row {
    display: flex;
    gap: 10px;
    align-items: center;
    margin-top: 14px;
  }

  .auth-status {
    color: var(--muted);
    font-size: 0.74rem;
    font-family: 'JetBrains Mono', monospace;
    margin-top: 10px;
    min-height: 1.2em;
  }

  .share-settings {
    display: grid;
    grid-template-columns: 1fr 160px;
    gap: 8px;
    margin-top: 10px;
  }

  .auth-pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 10px;
    border-radius: 999px;
    border: 1px solid var(--border);
    background: rgba(15,20,32,0.75);
    color: var(--muted);
    font-size: 0.7rem;
    font-family: 'JetBrains Mono', monospace;
  }

  .comment-meta {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
    color: var(--muted);
    font-size: 0.64rem;
    margin-top: 4px;
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
        <span class="auth-pill" id="authBadge" style="display:none">Auth required</span>
        <button class="chip-btn" type="button" id="logoutBtn" onclick="logout()" style="display:none">Logout</button>
        <span class="tag">v1.0.0</span>
      </div>
    </div>
  </header>

  <section class="workflow-picker" id="workflowPicker">
    <button class="workflow-card screenshare" type="button" onclick="openWorkflow('screenshare')">
      <div class="workflow-eyebrow">Capture first</div>
      <div class="workflow-title">Screenshare</div>
      <div class="workflow-copy">Record a walkthrough or load a meeting clip, then unlock trimming, timestamped notes, share links, and draft recovery after the video is ready.</div>
      <div class="workflow-features">
        <div class="workflow-feature">
          <strong>Native recorder</strong>
          <span>Capture the screen with optional system audio and microphone input.</span>
        </div>
        <div class="workflow-feature">
          <strong>Review workflow</strong>
          <span>Trim, annotate, and export a clean review session once capture finishes.</span>
        </div>
      </div>
      <div class="workflow-cta">Open screenshare workspace →</div>
    </button>

    <button class="workflow-card video2gif" type="button" onclick="openWorkflow('video2gif')">
      <div class="workflow-eyebrow">Convert next</div>
      <div class="workflow-title">Video 2 GIF</div>
      <div class="workflow-copy">Start with a source video, then reveal profile tuning, conversion controls, queued jobs, and server stats only after the upload is in place.</div>
      <div class="workflow-features">
        <div class="workflow-feature">
          <strong>Profile presets</strong>
          <span>Switch between fast, balanced, quality, tiny, HD, and custom output profiles.</span>
        </div>
        <div class="workflow-feature">
          <strong>Controlled output</strong>
          <span>Trim sections before conversion and monitor progress only when the asset is loaded.</span>
        </div>
      </div>
      <div class="workflow-cta">Open GIF workspace →</div>
    </button>
  </section>

  <div class="tabs" id="workflowTabs" style="display:none">
    <button class="tab-btn active" id="tabScreenShare" onclick="setTab('screenshare')">1) ScreenShare + Notes</button>
    <button class="tab-btn" id="tabGif" onclick="setTab('video2gif')">2) Video2GIF</button>
  </div>

  <main id="workspaceMain" style="display:none">
    <div id="sidebarCol">
      <div class="card">
        <div class="card-title" id="intakeCardTitle">Capture Workspace</div>
        <div class="mode-copy" id="modeCopy">Pick a workflow to get started.</div>

        <div id="screenRecorderBlock">
          <div class="recorder-row">
            <button class="chip-btn" type="button" id="recordStartBtn" onclick="startScreenRecording()">Start Recording</button>
            <button class="chip-btn" type="button" id="recordPauseBtn" onclick="toggleScreenRecordingPause()" disabled>Pause</button>
            <button class="chip-btn danger" type="button" id="recordStopBtn" onclick="stopScreenRecording()" disabled>Stop Recording</button>
            <span class="recorder-status" id="recorderStatus">
              <span class="rec-indicator"></span>
              <span id="recorderStatusText">Recorder idle</span>
            </span>
          </div>

          <div class="recorder-settings">
            <label class="inline-check"><input type="checkbox" id="recordSystemAudio" checked>System audio</label>
            <label class="inline-check"><input type="checkbox" id="recordMicrophone">Microphone</label>
            <div class="param-group">
              <label>Capture FPS</label>
              <select id="recordFrameRate">
                <option value="24">24 fps</option>
                <option value="30" selected>30 fps</option>
                <option value="60">60 fps</option>
              </select>
            </div>
            <div class="param-group">
              <label>Review Speed</label>
              <select id="playbackRate" onchange="setPlaybackRate(this.value)">
                <option value="0.75">0.75×</option>
                <option value="1" selected>1.00×</option>
                <option value="1.25">1.25×</option>
                <option value="1.5">1.50×</option>
                <option value="2">2.00×</option>
              </select>
            </div>
          </div>
        </div>

        <div class="dropzone" id="dropzone">
          <input type="file" id="fileInput" accept="video/*,.mkv,.avi,.flv,.wmv,.ts,.mts,.m2ts" />
          <div class="drop-icon" id="dropIcon">🎬</div>
          <div class="drop-text" id="dropText"><strong>Drop a video file here</strong> or click to browse</div>
          <div class="drop-formats" id="dropFormats">MP4 · MOV · MKV · AVI · WEBM · FLV · WMV · TS · 3GP</div>
        </div>

        <div class="file-preview" id="filePreview">
          <div class="file-icon">📹</div>
          <div class="file-info">
            <div class="file-name" id="previewName"></div>
            <div class="file-meta" id="previewMeta"></div>
          </div>
          <div class="remove-btn" onclick="clearFile()" title="Remove">✕</div>
        </div>

        <div class="share-panel" id="shareRow">
          <div class="share-label">Shareable Link</div>
          <div class="share-line">
            <button class="chip-btn" id="shareBtn" onclick="createShareLink()" disabled>Create Link</button>
            <input type="text" id="shareLink" readonly placeholder="Click Create Link to generate URL" />
            <button class="chip-btn" type="button" onclick="copyShareLink()">Copy</button>
          </div>
          <div class="share-settings">
            <input type="text" id="shareAuthor" placeholder="Reviewer / author name" />
            <select id="shareExpiryHours">
              <option value="24">24h</option>
              <option value="72">72h</option>
              <option value="168" selected>7 days</option>
              <option value="336">14 days</option>
              <option value="720">30 days</option>
            </select>
          </div>
        </div>

        <div class="draft-panel" id="draftPanel">
          <div class="draft-head">
            <div class="share-label" style="margin:0">Draft Recovery</div>
            <span class="tag" id="draftTag">local</span>
          </div>
          <div class="draft-copy" id="draftStatus">No local draft yet. Cuts, notes, and review settings will autosave in this browser.</div>
          <div class="draft-actions">
            <button class="chip-btn" type="button" id="restoreDraftBtn" onclick="restoreDraftToCurrentFile()" disabled>Restore Draft</button>
            <button class="chip-btn danger" type="button" id="discardDraftBtn" onclick="discardDraft()">Discard Draft</button>
          </div>
        </div>
        <button class="save-video-wide" type="button" id="saveScreenBtn" onclick="saveEditedVideo()" disabled>Save Video</button>
      </div>

      <div class="card" style="margin-top:16px" id="gifCard">
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
          <div class="progress-steps" id="progressSteps"></div>
        </div>
        <div class="result-panel" id="resultPanel">
          <div class="result-head">
            <div>
              <div class="result-title" id="resultTitle">Result ready</div>
              <div class="result-meta" id="resultMeta"></div>
            </div>
          </div>
          <div class="result-preview" id="resultPreview"></div>
          <div class="result-actions">
            <a class="result-link" id="resultViewLink" href="#" target="_blank" rel="noopener">Open Result</a>
            <a class="result-link download" id="resultDownloadLink" href="#">Download</a>
          </div>
        </div>
      </div>
    </div>

    <div id="mainCol">
      <div class="card editor-card" id="editorCard">
        <div class="card-title">Video Editor</div>
        <div class="editor-wrap" id="editorWrap">
          <div class="comment-layout">
            <div>
              <div class="video-stage">
                <video id="editorVideo" class="editor-video" controls preload="metadata"></video>
                <div id="videoOverlay" class="video-overlay"></div>
              </div>

              <div class="editor-toolbar">
                <button class="chip-btn" type="button" onclick="seekRelative(-5)">−5s</button>
                <button class="chip-btn" type="button" onclick="seekRelative(5)">+5s</button>
                <button class="chip-btn" type="button" onclick="markCutStart()">Mark Cut Start</button>
                <button class="chip-btn" type="button" onclick="markCutEnd()">Mark Cut End</button>
                <button class="chip-btn warn" type="button" onclick="cutMarkedRange()">Cut Marked Range</button>
                <button class="chip-btn" type="button" id="loopToggleBtn" onclick="toggleLoopSelection()">Loop Cuts Off</button>
                <button class="chip-btn" type="button" id="undoBtn" onclick="undoHistory()" disabled>Undo</button>
                <button class="chip-btn" type="button" id="redoBtn" onclick="redoHistory()" disabled>Redo</button>
                <button class="chip-btn danger" type="button" onclick="resetCuts()">Reset Cuts</button>
                <span class="editor-time" id="editorNow">00:00.00 / 00:00.00</span>
              </div>
            </div>

            <div class="comment-sidebar">
              <div style="font-size:0.72rem;color:var(--muted);text-transform:uppercase;letter-spacing:0.08em">Notes / Comments</div>
              <div class="comment-tools" style="grid-template-columns:1fr">
                <input type="text" id="commentAuthor" placeholder="Comment author" />
              </div>
              <div class="comment-tools">
                <input type="text" id="commentText" placeholder="Comment text..." />
                <button class="chip-btn warn" type="button" onclick="toggleCommentCapture()">Pick Point</button>
              </div>
              <div class="comment-actions">
                <button class="chip-btn" type="button" onclick="addCommentAtCurrent()">Add at Current Time</button>
                <button class="chip-btn" type="button" onclick="updateActiveComment()">Update Selected</button>
                <button class="chip-btn" type="button" onclick="toggleCommentResolved()">Resolve / Reopen</button>
                <button class="chip-btn" type="button" onclick="copyNotesReport()">Copy Report</button>
                <button class="chip-btn danger" type="button" onclick="deleteActiveComment()">Delete Selected</button>
                <button class="chip-btn danger" type="button" onclick="clearComments()">Clear</button>
              </div>
              <div class="comment-summary" id="commentSummary">0 notes</div>
              <div class="comment-list" id="commentList"></div>
            </div>
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
            <span>Cut Ranges (Multi-Select) <span id="cutRemovedLabel" style="color:#fca5a5">· Removed 00:00.00</span></span>
            <button class="chip-btn danger" type="button" onclick="resetCuts()">Reset All</button>
          </div>
          <div class="segment-list" id="segmentList"></div>
          <div class="editor-summary" id="editorSummary"></div>
          <div class="shortcut-note">Shortcuts: Space play/pause, J/L seek, I/O mark cut bounds, K add cut, N focus note, Ctrl/Cmd+Z undo, Shift+Ctrl/Cmd+Z redo.</div>
        </div>
      </div>

      <div style="display:flex;flex-direction:column;gap:16px" id="rightPanel">
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
    </div>
  </main>
  <div class="bottom-save-wrap" id="bottomSaveWrap">
    <button class="save-video-wide" type="button" id="bottomSaveBtn" onclick="saveEditedVideo()" disabled>Save Video</button>
  </div>
</div>

<div id="toast"></div>
<div class="auth-gate" id="authGate">
  <div class="auth-card">
    <div class="card-title" style="margin-bottom:0">Workspace Login</div>
    <div class="auth-copy">This workspace is password protected. Shared review links can stay public, but creating or editing content requires signing in.</div>
    <input type="password" id="authPassword" placeholder="Enter workspace password" />
    <div class="auth-row">
      <button class="convert-btn" type="button" id="loginBtn" onclick="login()">Unlock Workspace</button>
    </div>
    <div class="auth-status" id="authStatusText"></div>
  </div>
</div>

<script>
const API = '/api/v1';
const APP_DRAFT_KEY = 'video2gif:draft:v2';
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
let microphoneStream = null;
let mediaRecorder = null;
let recordedChunks = [];
let recorderTimer = null;
let recorderStartedAt = 0;
let recorderElapsedMs = 0;
let activeTab = 'screenshare';
let workflowSelected = false;
let comments = [];
let commentCaptureMode = false;
let pendingCommentPoint = null;
let activeCommentId = '';
let pausedCommentIDs = new Set();
let loopSelectionEnabled = false;
let historyStack = [];
let redoStack = [];
let suppressHistory = false;
let draftCache = null;
let draftSaveTimer = null;
let draftCacheLoaded = false;
let pendingDraftRestore = false;
let authEnabled = false;
let isAuthenticated = false;
let activeJobID = '';
let activeJobKind = '';
let activeProgressStep = '';

// ── Init ──────────────────────────────────────────────────────────────────
document.addEventListener('DOMContentLoaded', async () => {
  loadDraftCache();
  setupDragDrop();
  setupKeyboardShortcuts();
  bindDraftAwareInputs();
  updateHistoryButtons();
  updateDraftStatus();
  document.getElementById('authPassword').addEventListener('keydown', e => {
    if (e.key === 'Enter') {
      e.preventDefault();
      login();
    }
  });
  const sharedView = !!new URLSearchParams(window.location.search).get('share');
  await bootstrapAuth(sharedView);
  if (sharedView) {
    await maybeLoadSharedSession();
  }
  checkHealth();
  setInterval(checkHealth, 15000);
  setInterval(() => {
    if (!authEnabled || isAuthenticated) {
      loadJobs();
    }
  }, 5000);
});

async function bootstrapAuth(sharedView) {
  try {
    const r = await fetch(API + '/auth/status', { credentials: 'same-origin' });
    const status = await r.json();
    authEnabled = !!status.enabled;
    isAuthenticated = !!status.authenticated;
    updateAuthUI(status, sharedView);
    if (!authEnabled || isAuthenticated) {
      await loadPrivateWorkspace();
    }
  } catch (e) {
    if (!sharedView) {
      setAuthMessage('Authentication status check failed', true);
    }
  }
}

async function loadPrivateWorkspace() {
  await loadProfiles();
  await loadJobs();
  await loadConfig();
  if (draftCache?.active_tab) {
    activeTab = draftCache.active_tab;
  }
  refreshWorkspaceUI();
}

async function apiFetch(url, options) {
  const response = await fetch(url, {
    credentials: 'same-origin',
    ...(options || {}),
  });
  if (response.status === 401) {
    handleUnauthorized();
  }
  return response;
}

async function uploadFormWithProgress(url, form, hooks) {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open('POST', url, true);
    xhr.withCredentials = true;
    xhr.responseType = 'json';
    xhr.upload.addEventListener('progress', event => {
      if (!hooks?.onProgress || !event.lengthComputable || !event.total) return;
      hooks.onProgress(event.loaded / event.total, event.loaded, event.total);
    });
    xhr.upload.addEventListener('loadend', () => {
      hooks?.onUploadComplete?.();
    });
    xhr.onload = () => {
      if (xhr.status === 401) handleUnauthorized();
      let data = xhr.response;
      if (!data || typeof data === 'string') {
        try {
          data = JSON.parse(xhr.responseText || '{}');
        } catch {
          data = {};
        }
      }
      resolve({
        ok: xhr.status >= 200 && xhr.status < 300,
        status: xhr.status,
        data: data || {},
      });
    };
    xhr.onloadstart = () => {
      hooks?.onStart?.();
    };
    xhr.onerror = () => reject(new Error('Network error'));
    xhr.onabort = () => reject(new Error('Upload canceled'));
    xhr.send(form);
  });
}

async function login() {
  const input = document.getElementById('authPassword');
  const button = document.getElementById('loginBtn');
  const password = (input.value || '').trim();
  if (!password) {
    setAuthMessage('Enter the workspace password', true);
    return;
  }
  button.disabled = true;
  button.textContent = 'Unlocking...';
  try {
    const r = await apiFetch(API + '/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password }),
    });
    const d = await r.json();
    if (!r.ok) throw new Error(d.error || 'Login failed');
    authEnabled = !!d.enabled;
    isAuthenticated = !!d.authenticated;
    input.value = '';
    setAuthMessage('');
    updateAuthUI(d, false);
    await loadPrivateWorkspace();
    toast('Workspace unlocked', 'success');
  } catch (e) {
    setAuthMessage(e.message || 'Login failed', true);
  } finally {
    button.disabled = false;
    button.textContent = 'Unlock Workspace';
  }
}

async function logout() {
  await apiFetch(API + '/auth/logout', { method: 'POST' });
  handleUnauthorized();
}

function handleUnauthorized() {
  isAuthenticated = false;
  updateAuthUI({ enabled: authEnabled, authenticated: false }, isShareView());
}

function updateAuthUI(status, sharedView) {
  authEnabled = !!status.enabled;
  isAuthenticated = !!status.authenticated;
  const gate = document.getElementById('authGate');
  const badge = document.getElementById('authBadge');
  const logoutBtn = document.getElementById('logoutBtn');
  if (badge) {
    badge.style.display = authEnabled ? 'inline-flex' : 'none';
    badge.textContent = isAuthenticated ? 'Authenticated' : 'Auth required';
  }
  if (logoutBtn) {
    logoutBtn.style.display = authEnabled && isAuthenticated ? 'inline-flex' : 'none';
  }
  if (!gate) return;
  gate.classList.toggle('visible', authEnabled && !isAuthenticated && !sharedView);
  if (gate.classList.contains('visible')) {
    setTimeout(() => document.getElementById('authPassword').focus(), 30);
  }
}

function setAuthMessage(message, isError) {
  const el = document.getElementById('authStatusText');
  if (!el) return;
  el.textContent = message || '';
  el.style.color = isError ? 'var(--danger)' : 'var(--muted)';
}

function isShareView() {
  return !!new URLSearchParams(window.location.search).get('share');
}

// ── Health ────────────────────────────────────────────────────────────────
async function checkHealth() {
  try {
    const r = await apiFetch(API + '/health');
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
    const r = await apiFetch(API + '/profiles');
    profiles = await r.json();
    renderProfiles();
    applyProfile('balanced');
    maybeApplyDraftSettings();
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
  queueDraftSave();
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

// ── Tabs ──────────────────────────────────────────────────────────────────
function openWorkflow(name) {
  workflowSelected = true;
  setTab(name);
}

function refreshWorkspaceUI() {
  const shared = document.body.classList.contains('shared-mode');
  const hasVideo = !!selectedFile || shared;
  const showPicker = !shared && !workflowSelected && !selectedFile;
  const isScreen = activeTab === 'screenshare';
  const picker = document.getElementById('workflowPicker');
  const tabs = document.getElementById('workflowTabs');
  const main = document.getElementById('workspaceMain');
  const editorCard = document.getElementById('editorCard');
  const gifCard = document.getElementById('gifCard');
  const rightPanel = document.getElementById('rightPanel');
  const shareRow = document.getElementById('shareRow');
  const draftPanel = document.getElementById('draftPanel');
  const saveBtn = document.getElementById('saveBtn');
  const saveScreenBtn = document.getElementById('saveScreenBtn');
  const bottomSaveWrap = document.getElementById('bottomSaveWrap');
  const screenRecorderBlock = document.getElementById('screenRecorderBlock');
  const intakeTitle = document.getElementById('intakeCardTitle');
  const modeCopy = document.getElementById('modeCopy');
  const dropIcon = document.getElementById('dropIcon');
  const dropText = document.getElementById('dropText');
  const dropFormats = document.getElementById('dropFormats');
  const tabScreenShare = document.getElementById('tabScreenShare');
  const tabGif = document.getElementById('tabGif');
  const convertBtn = document.getElementById('convertBtn');

  if (picker) picker.style.display = showPicker ? 'grid' : 'none';
  if (tabs) tabs.style.display = showPicker || shared ? 'none' : 'flex';
  if (main) {
    main.style.display = showPicker ? 'none' : 'grid';
    main.classList.toggle('setup-only', !shared && !hasVideo);
  }
  if (tabScreenShare) tabScreenShare.classList.toggle('active', isScreen);
  if (tabGif) tabGif.classList.toggle('active', !isScreen);

  if (screenRecorderBlock) screenRecorderBlock.style.display = isScreen ? 'block' : 'none';
  if (editorCard) editorCard.style.display = hasVideo ? 'block' : 'none';
  if (gifCard) gifCard.style.display = !isScreen && hasVideo ? 'block' : 'none';
  if (rightPanel) rightPanel.style.display = !isScreen && hasVideo ? 'flex' : 'none';
  if (shareRow) shareRow.style.display = isScreen && hasVideo ? 'block' : 'none';
  if (draftPanel) draftPanel.style.display = hasVideo ? 'block' : 'none';
  if (convertBtn) convertBtn.style.display = isScreen ? 'none' : 'block';
  if (saveBtn) saveBtn.style.display = !isScreen ? 'block' : 'none';
  if (saveScreenBtn) saveScreenBtn.style.display = isScreen && hasVideo ? 'block' : 'none';
  if (bottomSaveWrap) bottomSaveWrap.style.display = selectedFile ? 'flex' : 'none';

  if (intakeTitle) {
    intakeTitle.textContent = isScreen ? (hasVideo ? 'ScreenShare Review' : 'ScreenShare Capture') : (hasVideo ? 'Video 2 GIF Setup' : 'Video 2 GIF Intake');
  }
  if (modeCopy) {
    modeCopy.textContent = isScreen
      ? (hasVideo
        ? 'Your recording is ready. Trimming, notes, sharing, draft recovery, and clean video export are now available.'
        : 'Start a screen recording or upload an existing capture. Review, sharing, and recovery tools stay hidden until a video is ready.')
      : (hasVideo
        ? 'Your source video is loaded. Profiles, fine-tuning, jobs, config, and conversion controls are now available.'
        : 'Upload a source video first. GIF tuning, jobs, and server-side conversion controls appear only after the file is loaded.');
  }
  if (dropIcon) dropIcon.textContent = isScreen ? '🖥️' : '🎬';
  if (dropText) {
    dropText.innerHTML = isScreen
      ? '<strong>Drop a screen recording here</strong> or click to browse'
      : '<strong>Drop a video file here</strong> or click to browse';
  }
  if (dropFormats) {
    dropFormats.textContent = isScreen
      ? 'WEBM · MP4 · MOV · MKV · AVI · WMV · TS'
      : 'MP4 · MOV · MKV · AVI · WEBM · FLV · WMV · TS · 3GP';
  }
}

function setTab(name) {
  workflowSelected = true;
  activeTab = name;
  refreshWorkspaceUI();
  queueDraftSave();
}

async function maybeLoadSharedSession() {
  const id = new URLSearchParams(window.location.search).get('share');
  if (!id) return;
  try {
    const r = await apiFetch(API + '/share/' + id);
    if (!r.ok) throw new Error('share not found');
    const d = await r.json();
    workflowSelected = true;
    setTab('screenshare');
    activeCommentId = '';
    document.getElementById('commentText').value = '';
	comments = (d.comments || []).map(c => ({
		id: c.id || randomID(),
		time: c.time || 0,
		x: c.x || 0.5,
		y: c.y || 0.5,
		text: c.text || '',
		status: c.status === 'resolved' ? 'resolved' : 'open',
		author: c.author || '',
		created_at: c.created_at || '',
	}));
    cutRanges = (d.cut_ranges || []).map(c => ({ start: c.start || 0, end: c.end || 0 }));
	const wrap = document.getElementById('editorWrap');
	const video = document.getElementById('editorVideo');
	document.body.classList.add('shared-mode');
    refreshWorkspaceUI();
	wrap.style.display = 'block';
	document.getElementById('shareAuthor').value = d.created_by || '';
	document.getElementById('previewName').textContent = d.file_name || 'shared-video';
	document.getElementById('previewMeta').textContent = 'Shared by ' + (d.created_by || 'unknown') + ' · expires ' + formatDateTime(d.expires_at);
    document.getElementById('filePreview').style.display = 'flex';
    video.src = d.video_url;
    document.getElementById('convertBtn').disabled = true;
    document.getElementById('saveBtn').disabled = true;
    document.getElementById('saveScreenBtn').disabled = true;
    document.getElementById('bottomSaveBtn').disabled = true;
    document.getElementById('shareBtn').disabled = true;
    document.getElementById('recordStartBtn').disabled = true;
    document.getElementById('recordPauseBtn').disabled = true;
    document.getElementById('recordStopBtn').disabled = true;
    document.getElementById('restoreDraftBtn').disabled = true;
    renderSegments();
    renderCommentDots();
    renderComments();
    renderEditorSummary();
    toast('Loaded shared session', 'success');
  } catch (e) {
    document.body.classList.remove('shared-mode');
    refreshWorkspaceUI();
    toast('Could not load shared session: ' + e.message, 'error');
  }
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
    const includeSystemAudio = document.getElementById('recordSystemAudio').checked;
    const includeMicrophone = document.getElementById('recordMicrophone').checked;
    const captureFPS = parseInt(document.getElementById('recordFrameRate').value, 10) || 30;
    recorderStream = await getDisplayMedia({
      video: { frameRate: captureFPS },
      audio: includeSystemAudio,
    });
    if (includeMicrophone && navigator.mediaDevices && navigator.mediaDevices.getUserMedia) {
      try {
        microphoneStream = await navigator.mediaDevices.getUserMedia({ audio: true });
      } catch (micErr) {
        toast('Microphone permission denied, continuing without mic', 'error');
      }
    }
    if (microphoneStream && microphoneStream.getAudioTracks().length) {
      const merged = new MediaStream();
      recorderStream.getVideoTracks().forEach(track => merged.addTrack(track));
      recorderStream.getAudioTracks().forEach(track => merged.addTrack(track));
      microphoneStream.getAudioTracks().forEach(track => merged.addTrack(track));
      recorderStream = merged;
    }
    const mimeType = pickRecorderMimeType();
    mediaRecorder = mimeType ? new MediaRecorder(recorderStream, { mimeType }) : new MediaRecorder(recorderStream);
    recordedChunks = [];
    recorderElapsedMs = 0;

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
    queueDraftSave();
    toast('Screen recording started', 'success');
  } catch (err) {
    setRecorderStatus('Recorder idle', false);
    toggleRecorderButtons(false);
    stopRecorderTracks();
    toast('Could not start recording: ' + err.message, 'error');
  }
}

function toggleScreenRecordingPause() {
  if (!mediaRecorder) return;
  if (mediaRecorder.state === 'recording') {
    recorderElapsedMs += Math.max(0, Date.now() - recorderStartedAt);
    mediaRecorder.pause();
    stopRecorderTimer();
    setRecorderStatus('Recording paused', false);
  } else if (mediaRecorder.state === 'paused') {
    mediaRecorder.resume();
    recorderStartedAt = Date.now();
    startRecorderTimer();
    setRecorderStatus('REC ' + formatClock(Math.floor(getRecordedDurationMs() / 1000)), true);
  }
  toggleRecorderButtons(mediaRecorder.state === 'recording', mediaRecorder.state === 'paused');
}

function stopScreenRecording() {
  if (mediaRecorder && (mediaRecorder.state === 'recording' || mediaRecorder.state === 'paused')) {
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
  if (recorderStream) recorderStream.getTracks().forEach(t => t.stop());
  recorderStream = null;
  if (microphoneStream) microphoneStream.getTracks().forEach(t => t.stop());
  microphoneStream = null;
  recorderElapsedMs = 0;
}

function toggleRecorderButtons(isRecording, isPaused) {
  document.getElementById('recordStartBtn').disabled = isRecording || isPaused;
  document.getElementById('recordPauseBtn').disabled = !isRecording && !isPaused;
  document.getElementById('recordPauseBtn').textContent = isPaused ? 'Resume' : 'Pause';
  document.getElementById('recordStopBtn').disabled = !isRecording && !isPaused;
}

function setRecorderStatus(text, recording) {
  const el = document.getElementById('recorderStatus');
  const textEl = document.getElementById('recorderStatusText');
  if (textEl) textEl.textContent = text;
  el.classList.toggle('recording', !!recording);
}

function startRecorderTimer() {
  stopRecorderTimer();
  recorderTimer = setInterval(() => {
    const sec = Math.floor(getRecordedDurationMs() / 1000);
    setRecorderStatus('REC ' + formatClock(sec), true);
  }, 200);
}

function stopRecorderTimer() {
  if (!recorderTimer) return;
  clearInterval(recorderTimer);
  recorderTimer = null;
}

function getRecordedDurationMs() {
  if (mediaRecorder && mediaRecorder.state === 'recording') {
    return Math.max(0, recorderElapsedMs + (Date.now() - recorderStartedAt));
  }
  return Math.max(0, recorderElapsedMs);
}

// ── File handling ─────────────────────────────────────────────────────────
function setupDragDrop() {
  const dz = document.getElementById('dropzone');
  const input = document.getElementById('fileInput');
  const scrub = document.getElementById('timelineScrub');
  const video = document.getElementById('editorVideo');
  const track = document.getElementById('timelineTrack');
  const overlay = document.getElementById('videoOverlay');

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
    if (!(document.body.classList.contains('shared-mode') && !selectedFile)) {
      cutRanges = [];
      comments = [];
      pendingCommentPoint = null;
      activeCommentId = '';
      pausedCommentIDs = new Set();
      markStart = null;
      markEnd = null;
      document.getElementById('commentText').value = '';
    }
    renderTimeline();
    renderSegments();
    renderCommentDots();
    renderComments();
    renderEditorSummary();
    updateEditorHUD(0);
    ensurePlayableDuration(video);
    if (pendingDraftRestore && selectedFile) {
      pendingDraftRestore = false;
      maybeRestoreDraftForFile(selectedFile);
    }
    pushHistorySnapshot();
  });

  video.addEventListener('durationchange', () => {
    if (!videoDuration && Number.isFinite(video.duration) && video.duration > 0) {
      setEditorDuration(video.duration);
      renderTimeline();
      renderSegments();
      renderEditorSummary();
      updateEditorHUD(video.currentTime || 0);
    }
  });

  video.addEventListener('error', () => {
    toast('Video preview failed to load. You can still trim if duration probe succeeds.', 'error');
  });

  video.addEventListener('timeupdate', () => {
    if (!videoDuration) return;
    maybePauseForComment(video.currentTime);
    if (loopSelectionEnabled) {
      const activeCut = cutRanges.find(c => video.currentTime >= c.end - 0.03 && video.currentTime <= c.end + 0.25);
      if (activeCut) {
        video.currentTime = activeCut.start;
        video.play().catch(() => {});
        return;
      }
    }
    if (cutPreview && video.currentTime >= cutPreview.end) {
      video.pause();
      stopCutPreview();
    }
    scrub.value = video.currentTime.toFixed(2);
    updateEditorHUD(video.currentTime);
    renderCommentDots();
  });

  video.addEventListener('seeked', () => {
    pausedCommentIDs = new Set([...pausedCommentIDs].filter(id => {
      const c = comments.find(x => x.id === id);
      return c && c.time < video.currentTime;
    }));
  });

  overlay.addEventListener('click', (e) => {
    if (!commentCaptureMode) return;
    const box = getVideoContentBoxInOverlay();
    if (!box.width || !box.height) return;
    const rect = overlay.getBoundingClientRect();
    const px = e.clientX - rect.left - box.left;
    const py = e.clientY - rect.top - box.top;
    pendingCommentPoint = {
      x: clamp((px / box.width), 0, 1),
      y: clamp((py / box.height), 0, 1),
    };
    overlay.classList.remove('capture');
    commentCaptureMode = false;
    toast('Point selected. Click "Add at Current Time".', 'success');
    renderCommentDots();
  });

  document.addEventListener('mousemove', handleCutDrag);
  document.addEventListener('mouseup', endCutDrag);
  track.addEventListener('mouseleave', endCutDrag);
}

function handleFile(file) {
  selectedFile = file;
  activeJobID = '';
  activeJobKind = '';
  activeProgressStep = '';
  clearInterval(pollInterval);
  workflowSelected = true;
  comments = [];
  pendingCommentPoint = null;
  activeCommentId = '';
  pausedCommentIDs = new Set();
  document.getElementById('previewName').textContent = file.name;
  document.getElementById('previewMeta').textContent =
    formatBytes(file.size) + ' · ' + (file.type || 'video');
  document.getElementById('filePreview').style.display = 'flex';
  document.getElementById('shareLink').value = '';
  document.getElementById('convertBtn').disabled = false;
  document.getElementById('saveBtn').disabled = false;
  document.getElementById('saveScreenBtn').disabled = false;
  document.getElementById('bottomSaveBtn').disabled = false;
  document.getElementById('shareBtn').disabled = false;
  document.getElementById('restoreDraftBtn').disabled = !hasRestorableDraftForFile(file);
  clearResultPanel();
  initEditor(file);
  resetHistory();
  pendingDraftRestore = hasRestorableDraftForFile(file);
  refreshWorkspaceUI();
  updateDraftStatus();
}

function clearFile() {
  stopCutPreview();
  selectedFile = null;
  activeJobID = '';
  activeJobKind = '';
  activeProgressStep = '';
  clearInterval(pollInterval);
  document.getElementById('fileInput').value = '';
  document.getElementById('filePreview').style.display = 'none';
  document.getElementById('shareLink').value = '';
  document.getElementById('convertBtn').disabled = true;
  document.getElementById('saveBtn').disabled = true;
  document.getElementById('saveScreenBtn').disabled = true;
  document.getElementById('bottomSaveBtn').disabled = true;
  document.getElementById('shareBtn').disabled = true;
  document.getElementById('restoreDraftBtn').disabled = !draftCache;
  document.getElementById('progressWrap').style.display = 'none';
  document.getElementById('editorWrap').style.display = 'none';
  document.getElementById('segmentList').innerHTML = '';
  document.getElementById('editorSummary').innerHTML = '';
  document.getElementById('cutRemovedLabel').textContent = '· Removed 00:00.00';
  cutRanges = [];
  comments = [];
  pendingCommentPoint = null;
  activeCommentId = '';
  pausedCommentIDs = new Set();
  pendingDraftRestore = false;
  markStart = null;
  markEnd = null;
  videoDuration = 0;
  document.getElementById('commentText').value = '';
  renderCommentDots();
  renderComments();
  clearResultPanel();
  if (editorObjectURL) {
    URL.revokeObjectURL(editorObjectURL);
    editorObjectURL = '';
  }
  resetHistory();
  updateHistoryButtons();
  refreshWorkspaceUI();
  updateDraftStatus();
}

function initEditor(file) {
  const wrap = document.getElementById('editorWrap');
  const video = document.getElementById('editorVideo');
  if (editorObjectURL) URL.revokeObjectURL(editorObjectURL);
  editorObjectURL = URL.createObjectURL(file);
  video.src = editorObjectURL;
  video.playbackRate = parseFloat(document.getElementById('playbackRate').value) || 1;
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
  renderEditorSummary();
}

async function probeVideoDuration(file) {
  try {
    const form = new FormData();
    form.append('video', file, file.name || 'recording.webm');
    const r = await apiFetch(API + '/probe', { method: 'POST', body: form });
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

function toggleCommentCapture() {
  const overlay = document.getElementById('videoOverlay');
  commentCaptureMode = !commentCaptureMode;
  overlay.classList.toggle('capture', commentCaptureMode);
  toast(commentCaptureMode ? 'Click on video to place note point' : 'Point selection cancelled', 'success');
}

function addCommentAtCurrent() {
  const video = document.getElementById('editorVideo');
  const text = (document.getElementById('commentText').value || '').trim();
  const author = (document.getElementById('commentAuthor').value || document.getElementById('shareAuthor').value || '').trim();
  if (!text) {
    toast('Enter a comment first', 'error');
    return;
  }
  if (!video || !Number.isFinite(video.currentTime)) {
    toast('Video is not ready', 'error');
    return;
  }
  const point = pendingCommentPoint || { x: 0.5, y: 0.5 };
  comments.push({
    id: randomID(),
    time: Number(video.currentTime.toFixed(3)),
    x: Number(point.x.toFixed(4)),
    y: Number(point.y.toFixed(4)),
    text,
    status: 'open',
    author,
    created_at: new Date().toISOString(),
  });
  comments.sort((a, b) => a.time - b.time);
  pendingCommentPoint = null;
  document.getElementById('commentText').value = '';
  pausedCommentIDs = new Set();
  renderComments();
  renderCommentDots();
  renderEditorSummary();
  pushHistorySnapshot();
  queueDraftSave();
}

function updateActiveComment() {
  const text = (document.getElementById('commentText').value || '').trim();
  const author = (document.getElementById('commentAuthor').value || document.getElementById('shareAuthor').value || '').trim();
  if (!activeCommentId) {
    toast('Select a comment to update', 'error');
    return;
  }
  if (!text) {
    toast('Enter updated comment text first', 'error');
    return;
  }
  const target = comments.find(c => c.id === activeCommentId);
  if (!target) return;
  target.text = text;
  target.author = author;
  renderComments();
  renderCommentDots();
  pushHistorySnapshot();
  queueDraftSave();
  toast('Comment updated', 'success');
}

function toggleCommentResolved(id) {
  const commentID = id || activeCommentId;
  if (!commentID) {
    toast('Select a comment first', 'error');
    return;
  }
  const target = comments.find(c => c.id === commentID);
  if (!target) return;
  target.status = target.status === 'resolved' ? 'open' : 'resolved';
  renderComments();
  renderCommentDots();
  renderEditorSummary();
  pushHistorySnapshot();
  queueDraftSave();
}

function deleteActiveComment(id) {
  const commentID = id || activeCommentId;
  if (!commentID) {
    toast('Select a comment first', 'error');
    return;
  }
  comments = comments.filter(c => c.id !== commentID);
  if (activeCommentId === commentID) {
    activeCommentId = '';
    document.getElementById('commentText').value = '';
  }
  pausedCommentIDs.delete(commentID);
  renderComments();
  renderCommentDots();
  renderEditorSummary();
  pushHistorySnapshot();
  queueDraftSave();
}

function clearComments() {
  comments = [];
  activeCommentId = '';
  pausedCommentIDs = new Set();
  document.getElementById('commentText').value = '';
  renderComments();
  renderCommentDots();
  renderEditorSummary();
  pushHistorySnapshot();
  queueDraftSave();
}

function renderComments() {
  const list = document.getElementById('commentList');
  const summary = document.getElementById('commentSummary');
  const openCount = comments.filter(c => c.status !== 'resolved').length;
  const resolvedCount = comments.length - openCount;
  summary.textContent = comments.length + ' notes · ' + openCount + ' open · ' + resolvedCount + ' resolved';
  if (!comments.length) {
    list.innerHTML = '<div class="segment-empty">No notes yet.</div>';
    return;
  }
  list.innerHTML = comments.map(c =>
    '<div class="comment-item' + (c.id === activeCommentId ? ' active' : '') + (c.status === 'resolved' ? ' resolved' : '') + '" onclick="jumpToComment(\'' + c.id + '\')">' +
      '<div class="comment-row">' +
        '<div class="comment-main">' +
          '<div>' + formatTime(c.time) + '</div>' +
          '<div class="comment-meta"><span>' + escHtml(c.author || 'anonymous') + '</span><span>' + formatDateTime(c.created_at) + '</span></div>' +
          '<div class="comment-text">' + escHtml(c.text) + '</div>' +
        '</div>' +
        '<span class="comment-badge' + (c.status === 'resolved' ? ' resolved' : '') + '">' + (c.status === 'resolved' ? 'Resolved' : 'Open') + '</span>' +
      '</div>' +
      '<div class="comment-actions">' +
        '<button class="comment-mini-btn" type="button" onclick="event.stopPropagation();jumpToComment(\'' + c.id + '\')">Jump</button>' +
        '<button class="comment-mini-btn" type="button" onclick="event.stopPropagation();toggleCommentResolved(\'' + c.id + '\')">' + (c.status === 'resolved' ? 'Reopen' : 'Resolve') + '</button>' +
        '<button class="comment-mini-btn" type="button" onclick="event.stopPropagation();deleteActiveComment(\'' + c.id + '\')">Delete</button>' +
      '</div>' +
    '</div>'
  ).join('');
}

function renderCommentDots() {
  const overlay = document.getElementById('videoOverlay');
  if (!overlay) return;
  const box = getVideoContentBoxInOverlay();
  const points = comments.map(c => {
    const leftPct = ((box.left + (c.x * box.width)) / box.overlayWidth) * 100;
    const topPct = ((box.top + (c.y * box.height)) / box.overlayHeight) * 100;
    return '<span class="comment-dot' + (c.id === activeCommentId ? ' active' : '') + (c.status === 'resolved' ? ' resolved' : '') + '" style="left:' + leftPct + '%;top:' + topPct + '%"></span>';
  }).join('');
  let pending = '';
  if (pendingCommentPoint) {
    const leftPct = ((box.left + (pendingCommentPoint.x * box.width)) / box.overlayWidth) * 100;
    const topPct = ((box.top + (pendingCommentPoint.y * box.height)) / box.overlayHeight) * 100;
    pending = '<span class="comment-dot active" style="left:' + leftPct + '%;top:' + topPct + '%"></span>';
  }
  overlay.innerHTML = points + pending;
}

function getVideoContentBoxInOverlay() {
  const overlay = document.getElementById('videoOverlay');
  const video = document.getElementById('editorVideo');
  const overlayWidth = Math.max(1, overlay?.clientWidth || 1);
  const overlayHeight = Math.max(1, overlay?.clientHeight || 1);
  const sourceWidth = Math.max(1, video?.videoWidth || overlayWidth);
  const sourceHeight = Math.max(1, video?.videoHeight || overlayHeight);
  const scale = Math.min(overlayWidth / sourceWidth, overlayHeight / sourceHeight);
  const width = sourceWidth * scale;
  const height = sourceHeight * scale;
  const left = (overlayWidth - width) / 2;
  const top = (overlayHeight - height) / 2;
  return { left, top, width, height, overlayWidth, overlayHeight };
}

function jumpToComment(id) {
  const c = comments.find(x => x.id === id);
  const video = document.getElementById('editorVideo');
  if (!c || !video) return;
  video.currentTime = c.time;
  activeCommentId = c.id;
  document.getElementById('commentText').value = c.text || '';
  document.getElementById('commentAuthor').value = c.author || '';
  renderComments();
  renderCommentDots();
}

function maybePauseForComment(currentTime) {
  const video = document.getElementById('editorVideo');
  if (!video || video.paused) return;
  for (const c of comments) {
    if (c.status === 'resolved') continue;
    if (pausedCommentIDs.has(c.id)) continue;
    if (currentTime >= c.time && currentTime <= c.time + 0.25) {
      pausedCommentIDs.add(c.id);
      activeCommentId = c.id;
      renderComments();
      renderCommentDots();
      video.pause();
      setTimeout(() => {
        if (!video.paused) return;
        video.play().catch(() => {});
      }, 1000);
      break;
    }
  }
}

async function createShareLink() {
  if (!selectedFile) {
    toast('Select or record a video first', 'error');
    return;
  }
  const btn = document.getElementById('shareBtn');
  btn.disabled = true;
  btn.textContent = 'Creating...';
  const form = new FormData();
  form.append('video', selectedFile, selectedFile.name || 'recording.webm');
  form.append('cut_ranges', JSON.stringify(cutRanges));
  form.append('comments', JSON.stringify(comments));
  form.append('created_by', (document.getElementById('shareAuthor').value || '').trim());
  form.append('expires_in_hours', document.getElementById('shareExpiryHours').value || '168');
  if (videoDuration > 0) {
    form.append('duration_hint', String(Number(videoDuration.toFixed(3))));
  }
  try {
    const r = await apiFetch(API + '/share', { method: 'POST', body: form });
    const d = await r.json();
    if (!r.ok) throw new Error(d.error || ('HTTP ' + r.status));
    document.getElementById('shareLink').value = d.share_url || '';
    if (d.share_url) {
      navigator.clipboard?.writeText(d.share_url).catch(() => {});
    }
    if (d.expires_at) {
      document.getElementById('previewMeta').textContent = document.getElementById('previewMeta').textContent.split(' · ')[0] + ' · share expires ' + formatDateTime(d.expires_at);
    }
    queueDraftSave();
    toast('Share link created and copied', 'success');
  } catch (e) {
    toast('Create share link failed: ' + e.message, 'error');
  } finally {
    btn.disabled = false;
    btn.textContent = 'Create Link';
  }
}

function copyShareLink() {
  const link = document.getElementById('shareLink').value.trim();
  if (!link) {
    toast('No share link available yet', 'error');
    return;
  }
  if (!navigator.clipboard || !navigator.clipboard.writeText) {
    toast('Clipboard API unavailable. Copy manually from the field.', 'error');
    return;
  }
  navigator.clipboard.writeText(link)
    .then(() => toast('Share link copied', 'success'))
    .catch(() => toast('Copy failed. You can copy manually from the field.', 'error'));
}

function copyNotesReport() {
  const report = buildNotesReport();
  if (!navigator.clipboard || !navigator.clipboard.writeText) {
    toast('Clipboard API unavailable for note export', 'error');
    return;
  }
  navigator.clipboard.writeText(report)
    .then(() => toast('Notes report copied', 'success'))
    .catch(() => toast('Copy failed. Report remains available in-page.', 'error'));
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
  renderEditorSummary();
  pushHistorySnapshot();
  queueDraftSave();
}

function resetCuts() {
  stopCutPreview();
  cutRanges = [];
  renderTimeline();
  renderSegments();
  renderEditorSummary();
  pushHistorySnapshot();
  queueDraftSave();
}

function removeCut(index) {
  if (index < 0 || index >= cutRanges.length) return;
  stopCutPreview();
  cutRanges.splice(index, 1);
  cutRanges = mergeRanges(cutRanges);
  renderTimeline();
  renderSegments();
  renderEditorSummary();
  pushHistorySnapshot();
  queueDraftSave();
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
  renderEditorSummary();
  pushHistorySnapshot();
  queueDraftSave();
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
  if (!track || !playhead) return;
  track.querySelectorAll('.timeline-seg,.timeline-cut').forEach(n => n.remove());
  if (!videoDuration) return;

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
      '<span class="segment-pill">Cut ' + formatTime(Math.max(0, s.end - s.start)) + '</span>' +
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
  renderEditorSummary();
  pushHistorySnapshot();
  queueDraftSave();
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

function seekRelative(delta) {
  const video = document.getElementById('editorVideo');
  if (!video || !videoDuration) return;
  video.currentTime = clamp((video.currentTime || 0) + delta, 0, videoDuration);
  updateEditorHUD(video.currentTime || 0);
}

function setPlaybackRate(value) {
  const video = document.getElementById('editorVideo');
  if (video) video.playbackRate = parseFloat(value) || 1;
  queueDraftSave();
}

function toggleLoopSelection() {
  loopSelectionEnabled = !loopSelectionEnabled;
  document.getElementById('loopToggleBtn').textContent = loopSelectionEnabled ? 'Loop Cuts On' : 'Loop Cuts Off';
}

function renderEditorSummary() {
  const el = document.getElementById('editorSummary');
  if (!el) return;
  const removedDuration = getTotalRemovedDuration();
  const keptDuration = Math.max(0, videoDuration - removedDuration);
  const openNotes = comments.filter(c => c.status !== 'resolved').length;
  const removedLabel = document.getElementById('cutRemovedLabel');
  if (removedLabel) {
    removedLabel.textContent = '· Removed ' + formatTime(removedDuration);
  }
  el.innerHTML = [
    '<span class="summary-pill">Duration ' + formatTime(videoDuration) + '</span>',
    '<span class="summary-pill">Removed ' + formatTime(removedDuration) + '</span>',
    '<span class="summary-pill">Kept ' + formatTime(keptDuration) + '</span>',
    '<span class="summary-pill">Cuts ' + cutRanges.length + '</span>',
    '<span class="summary-pill">Open Notes ' + openNotes + '</span>',
    '<span class="summary-pill">Resolved ' + (comments.length - openNotes) + '</span>',
  ].join('');
}

function captureEditorState() {
  return {
    cutRanges: cutRanges.map(c => ({ start: c.start, end: c.end })),
    comments: comments.map(c => ({
      id: c.id,
      time: c.time,
      x: c.x,
      y: c.y,
      text: c.text,
      status: c.status === 'resolved' ? 'resolved' : 'open',
      author: c.author || '',
      created_at: c.created_at || '',
    })),
    markStart,
    markEnd,
    activeCommentId,
  };
}

function getNormalizedCutPayload() {
  return mergeRanges(cutRanges).map(s => ({
    start: Number(s.start.toFixed(3)),
    end: Number(s.end.toFixed(3)),
  }));
}

function getTotalRemovedDuration() {
  return getNormalizedCutPayload().reduce((sum, cut) => sum + Math.max(0, cut.end - cut.start), 0);
}

function applyEditorState(state) {
  suppressHistory = true;
  cutRanges = (state.cutRanges || []).map(c => ({ start: c.start || 0, end: c.end || 0 }));
	comments = (state.comments || []).map(c => ({
		id: c.id || randomID(),
		time: c.time || 0,
		x: c.x || 0.5,
		y: c.y || 0.5,
		text: c.text || '',
		status: c.status === 'resolved' ? 'resolved' : 'open',
		author: c.author || '',
		created_at: c.created_at || '',
	}));
  markStart = Number.isFinite(state.markStart) ? state.markStart : null;
  markEnd = Number.isFinite(state.markEnd) ? state.markEnd : null;
  activeCommentId = state.activeCommentId || '';
  syncCutBounds();
  pausedCommentIDs = new Set();
  if (activeCommentId) {
    const selected = comments.find(c => c.id === activeCommentId);
    document.getElementById('commentText').value = selected ? selected.text : '';
  } else {
    document.getElementById('commentText').value = '';
  }
  renderTimeline();
  renderSegments();
  renderComments();
  renderCommentDots();
  renderEditorSummary();
  suppressHistory = false;
}

function pushHistorySnapshot() {
  if (suppressHistory || !selectedFile) {
    updateHistoryButtons();
    return;
  }
  const next = captureEditorState();
  const prev = historyStack[historyStack.length - 1];
  if (prev && JSON.stringify(prev) === JSON.stringify(next)) {
    updateHistoryButtons();
    return;
  }
  historyStack.push(next);
  if (historyStack.length > 80) historyStack.shift();
  redoStack = [];
  updateHistoryButtons();
}

function resetHistory() {
  historyStack = [];
  redoStack = [];
}

function updateHistoryButtons() {
  document.getElementById('undoBtn').disabled = historyStack.length < 2;
  document.getElementById('redoBtn').disabled = redoStack.length === 0;
}

function undoHistory() {
  if (historyStack.length < 2) return;
  const current = historyStack.pop();
  redoStack.push(current);
  applyEditorState(historyStack[historyStack.length - 1]);
  updateHistoryButtons();
  queueDraftSave();
}

function redoHistory() {
  if (!redoStack.length) return;
  const next = redoStack.pop();
  historyStack.push(next);
  applyEditorState(next);
  updateHistoryButtons();
  queueDraftSave();
}

function loadDraftCache() {
  draftCacheLoaded = true;
  try {
    const raw = localStorage.getItem(APP_DRAFT_KEY);
    draftCache = raw ? JSON.parse(raw) : null;
  } catch {
    draftCache = null;
  }
  updateDraftStatus();
}

function getSelectedFileSignature(file) {
  const target = file || selectedFile;
  if (!target) return '';
  return [target.name || '', target.size || 0, target.type || '', target.lastModified || 0].join('::');
}

function captureSettingsState() {
	return {
		activeProfile,
    fps: document.getElementById('fps').value,
    colors: document.getElementById('colors').value,
    width: document.getElementById('width').value,
    height: document.getElementById('height').value,
    dither: document.getElementById('dither').value,
    bayerScale: document.getElementById('bayerScale').value,
    speed: document.getElementById('speed').value,
    loop: document.getElementById('loop').value,
    optimizePalette: document.getElementById('optimizePalette').checked,
		recordSystemAudio: document.getElementById('recordSystemAudio').checked,
		recordMicrophone: document.getElementById('recordMicrophone').checked,
		recordFrameRate: document.getElementById('recordFrameRate').value,
		playbackRate: document.getElementById('playbackRate').value,
		shareAuthor: document.getElementById('shareAuthor').value,
		commentAuthor: document.getElementById('commentAuthor').value,
		shareExpiryHours: document.getElementById('shareExpiryHours').value,
	};
}

function applySettingsState(state) {
  if (!state) return;
  if (state.activeProfile && profiles[state.activeProfile]) {
    applyProfile(state.activeProfile);
  }
  const pairs = [
    ['fps', state.fps],
    ['colors', state.colors],
    ['width', state.width],
    ['height', state.height],
    ['dither', state.dither],
    ['bayerScale', state.bayerScale],
    ['speed', state.speed],
    ['loop', state.loop],
		['recordFrameRate', state.recordFrameRate],
		['playbackRate', state.playbackRate],
		['shareAuthor', state.shareAuthor],
		['commentAuthor', state.commentAuthor],
		['shareExpiryHours', state.shareExpiryHours],
	];
  pairs.forEach(([id, value]) => {
    if (value === undefined || value === null) return;
    const el = document.getElementById(id);
    if (el) el.value = value;
  });
  if (state.optimizePalette !== undefined) {
    document.getElementById('optimizePalette').checked = !!state.optimizePalette;
  }
  if (state.recordSystemAudio !== undefined) {
    document.getElementById('recordSystemAudio').checked = !!state.recordSystemAudio;
  }
  if (state.recordMicrophone !== undefined) {
    document.getElementById('recordMicrophone').checked = !!state.recordMicrophone;
  }
  syncSlider('fps', 'fpsVal');
  syncSlider('colors', 'colorsVal');
  syncSlider('bayerScale', 'bayerVal');
  syncSlider('speed', 'speedVal');
  setPlaybackRate(document.getElementById('playbackRate').value);
}

function queueDraftSave() {
  if (!draftCacheLoaded) return;
  clearTimeout(draftSaveTimer);
  draftSaveTimer = setTimeout(saveDraft, 120);
}

function saveDraft() {
  if (!draftCacheLoaded) return;
  const base = (!selectedFile && draftCache) ? draftCache : {};
  draftCache = {
    ...base,
    saved_at: new Date().toISOString(),
    active_tab: activeTab,
    file_signature: selectedFile ? getSelectedFileSignature() : (base.file_signature || ''),
    file_name: selectedFile ? selectedFile.name : (base.file_name || ''),
    video_duration: selectedFile ? Number(videoDuration.toFixed ? videoDuration.toFixed(3) : 0) : (base.video_duration || 0),
    settings: captureSettingsState(),
    editor: selectedFile ? captureEditorState() : (base.editor || captureEditorState()),
    share_link: document.getElementById('shareLink').value || '',
  };
  try {
    localStorage.setItem(APP_DRAFT_KEY, JSON.stringify(draftCache));
  } catch {}
  updateDraftStatus();
}

function hasRestorableDraftForFile(file) {
  return !!(draftCache && file && draftCache.file_signature && draftCache.file_signature === getSelectedFileSignature(file));
}

function maybeApplyDraftSettings() {
  if (draftCache && draftCache.settings) {
    applySettingsState(draftCache.settings);
    if (draftCache.active_tab) {
      activeTab = draftCache.active_tab;
    }
    refreshWorkspaceUI();
  }
}

function maybeRestoreDraftForFile(file) {
  if (!hasRestorableDraftForFile(file) || !draftCache?.editor) return;
  applyEditorState(draftCache.editor);
  document.getElementById('shareLink').value = draftCache.share_link || '';
  renderEditorSummary();
  toast('Local draft restored for this file', 'success');
}

function restoreDraftToCurrentFile() {
  if (!selectedFile) {
    toast('Open the matching file first', 'error');
    return;
  }
  if (!hasRestorableDraftForFile(selectedFile)) {
    toast('No matching draft for this file', 'error');
    return;
  }
  maybeRestoreDraftForFile(selectedFile);
  pushHistorySnapshot();
}

function discardDraft() {
  clearTimeout(draftSaveTimer);
  draftCache = null;
  try {
    localStorage.removeItem(APP_DRAFT_KEY);
  } catch {}
  updateDraftStatus();
}

function updateDraftStatus() {
  const status = document.getElementById('draftStatus');
  const restoreBtn = document.getElementById('restoreDraftBtn');
  if (!status || !restoreBtn) return;
  if (!draftCache) {
    status.textContent = 'No local draft yet. Cuts, notes, and review settings will autosave in this browser.';
    restoreBtn.disabled = true;
    return;
  }
  const savedAt = draftCache.saved_at ? new Date(draftCache.saved_at).toLocaleString() : 'unknown time';
  if (selectedFile && hasRestorableDraftForFile(selectedFile)) {
    status.textContent = 'Draft found for "' + (draftCache.file_name || selectedFile.name) + '" from ' + savedAt + '. Restore or continue editing.';
    restoreBtn.disabled = false;
    return;
  }
  status.textContent = 'Saved draft for "' + (draftCache.file_name || 'untitled session') + '" from ' + savedAt + '. Open the same file to restore cuts and notes.';
  restoreBtn.disabled = true;
}

function bindDraftAwareInputs() {
	['fps', 'colors', 'width', 'height', 'dither', 'bayerScale', 'speed', 'loop', 'optimizePalette', 'recordSystemAudio', 'recordMicrophone', 'recordFrameRate', 'playbackRate', 'cutStart', 'cutEnd', 'shareAuthor', 'commentAuthor', 'shareExpiryHours'].forEach(id => {
    const el = document.getElementById(id);
    if (!el) return;
    el.addEventListener('change', queueDraftSave);
    el.addEventListener('input', queueDraftSave);
  });
}

function setupKeyboardShortcuts() {
  document.addEventListener('keydown', event => {
    const tag = (event.target?.tagName || '').toLowerCase();
    const isTyping = tag === 'input' || tag === 'textarea' || tag === 'select' || event.target?.isContentEditable;
    const mod = event.metaKey || event.ctrlKey;

    if (mod && event.key.toLowerCase() === 'z') {
      event.preventDefault();
      if (event.shiftKey) redoHistory();
      else undoHistory();
      return;
    }

    if (isTyping || !selectedFile) return;

    if (event.code === 'Space') {
      event.preventDefault();
      const video = document.getElementById('editorVideo');
      if (!video) return;
      if (video.paused) video.play().catch(() => {});
      else video.pause();
    } else if (event.key.toLowerCase() === 'j') {
      event.preventDefault();
      seekRelative(-5);
    } else if (event.key.toLowerCase() === 'l') {
      event.preventDefault();
      seekRelative(5);
    } else if (event.key.toLowerCase() === 'i') {
      event.preventDefault();
      markCutStart();
    } else if (event.key.toLowerCase() === 'o') {
      event.preventDefault();
      markCutEnd();
    } else if (event.key.toLowerCase() === 'k') {
      event.preventDefault();
      cutMarkedRange();
    } else if (event.key.toLowerCase() === 'n') {
      event.preventDefault();
      document.getElementById('commentText').focus();
    }
  });
}

function buildNotesReport() {
  const title = selectedFile ? selectedFile.name : 'screen-session';
  const lines = ['Review Notes: ' + title];
  if (!comments.length) {
    lines.push('No notes recorded.');
    return lines.join('\n');
  }
  comments.forEach((c, index) => {
    const author = c.author ? ' @' + c.author : '';
    lines.push((index + 1) + '. [' + (c.status === 'resolved' ? 'resolved' : 'open') + '] ' + formatTime(c.time) + author + ' - ' + c.text);
  });
  return lines.join('\n');
}

// ── Conversion ────────────────────────────────────────────────────────────
async function saveEditedVideo() {
  if (!selectedFile) return;
  setActionButtonsBusy('save');
  clearResultPanel();
  activeJobKind = 'video';
  activeProgressStep = 'uploading';
  showProgress(true);
  setProgress(1, 'Uploading source video...', 'Preparing edited-video job', 'uploading', false, 1);

  const cutPayload = getNormalizedCutPayload();
  const form = new FormData();
  form.append('video', selectedFile, selectedFile.name || 'recording.webm');
  form.append('cut_ranges', JSON.stringify(cutPayload));
  if (videoDuration > 0) {
    form.append('duration_hint', String(Number(videoDuration.toFixed(3))));
  }

  try {
    const response = await uploadFormWithProgress(API + '/save-edited', form, {
      onProgress: (ratio, loaded, total) => {
        const pct = Math.max(1, Math.min(20, Math.round(ratio * 20)));
        setProgress(pct, 'Uploading source video...', formatBytes(loaded) + ' / ' + formatBytes(total), 'uploading', false, Math.max(1, Math.min(100, Math.round(ratio * 100))));
      },
      onUploadComplete: () => {
        setProgress(22, 'Upload complete', 'Server is validating the file and creating the edit job', 'setup', false, 5);
      },
    });
    if (!response.ok) {
      throw new Error(response.data.error || ('HTTP ' + response.status));
    }
    const job = response.data;
    setProgress(28, 'Queued...', 'Edited video job accepted by the server', 'queued');
    pollJob(job.id);
    toast('Edited video job queued. You’ll see the finished MP4 below when it completes.', 'success');
  } catch (e) {
    setProgress(0, 'Save failed', e.message || 'Unable to save edited video', activeProgressStep || 'uploading', true);
    setTimeout(() => showProgress(false), 1800);
    toast('Save failed: ' + e.message, 'error');
  } finally {
    if (!activeJobID) {
      setActionButtonsIdle();
    }
  }
}

async function startConvert() {
  if (!selectedFile) return;
  setActionButtonsBusy('convert');
  clearResultPanel();
  activeJobKind = 'gif';
  activeProgressStep = 'uploading';
  showProgress(true);
  setProgress(1, 'Uploading source video...', 'Preparing GIF job', 'uploading', false, 1);

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
    const response = await uploadFormWithProgress(API + '/convert', form, {
      onProgress: (ratio, loaded, total) => {
        const pct = Math.max(1, Math.min(20, Math.round(ratio * 20)));
        setProgress(pct, 'Uploading source video...', formatBytes(loaded) + ' / ' + formatBytes(total), 'uploading', false, Math.max(1, Math.min(100, Math.round(ratio * 100))));
      },
      onUploadComplete: () => {
        setProgress(22, 'Upload complete', 'Server is validating the file and creating the GIF job', 'setup', false, 5);
      },
    });
    const job = response.data;
    if (!response.ok) {
      throw new Error(job.error || ('HTTP ' + response.status));
    }
    setProgress(28, 'Queued...', 'GIF job accepted by the server', 'queued');
    pollJob(job.id);
    toast('Job submitted! ID: ' + job.id.slice(0,8) + '…', 'success');
  } catch (e) {
    setProgress(0, 'Upload failed', e.message || 'Unable to start conversion', activeProgressStep || 'uploading', true);
    setTimeout(() => showProgress(false), 1800);
    toast('Network error: ' + e.message, 'error');
  } finally {
    if (!activeJobID) {
      setActionButtonsIdle();
    }
  }
}

function pollJob(id) {
  clearInterval(pollInterval);
  activeJobID = id;

  const tick = async () => {
    try {
      const r = await apiFetch(API + '/jobs/' + id);
      const job = await r.json();
      if (!r.ok) throw new Error(job.error || ('HTTP ' + r.status));
      updateProgress(job);

      if (job.status === 'done') {
        clearInterval(pollInterval);
        activeJobID = '';
        const downloadKind = job.kind === 'video' ? 'Edited video' : 'GIF';
        setProgress(100, 'Complete!', downloadKind + ' ready for download', 'complete', false, 100);
        showJobResult(job);
        toast(downloadKind + ' complete!', 'success');
        setActionButtonsIdle();
        loadJobs();
      } else if (job.status === 'failed') {
        clearInterval(pollInterval);
        activeJobID = '';
        setProgress(jobProgressPercent(job), 'Failed', '❌ ' + (job.error || 'Unknown error'), deriveJobStep(job), true, deriveStagePercent(job, deriveJobStep(job)));
        toast((job.kind === 'video' ? 'Video export failed: ' : 'Conversion failed: ') + (job.error || ''), 'error');
        setActionButtonsIdle();
        loadJobs();
      }
    } catch (e) {
      if (!activeJobID) return;
      setProgress(24, 'Checking job status...', 'Retrying after a temporary network issue', activeProgressStep || 'queued');
    }
  };

  tick();
  pollInterval = setInterval(tick, 1200);
}

function updateProgress(job) {
  const statusLabels = {
    queued: 'Queued...',
    running: job.stage || 'Processing with FFmpeg...',
    done: 'Complete!',
    failed: 'Failed',
  };
  const detail = job.detail || (job.status === 'queued' ? 'Waiting for an available worker...' : '');
  activeJobKind = job.kind || activeJobKind;
  const stepKey = deriveJobStep(job);
  setProgress(jobProgressPercent(job), statusLabels[job.status] || job.status, detail, stepKey, false, deriveStagePercent(job, stepKey));
  loadJobs();
}

function jobProgressPercent(job) {
  if (!job) return 0;
  if (job.status === 'done') return 100;
  if (job.status === 'failed') {
    const prior = Number(job.progress || 0);
    return Math.max(0, Math.min(99, Math.round(20 + (prior * 80))));
  }
  if (job.status === 'queued') return 20;
  const raw = Number(job.progress || 0);
  return Math.max(21, Math.min(99, Math.round(20 + (raw * 80))));
}

function setActionButtonsBusy(mode) {
  const convertBtn = document.getElementById('convertBtn');
  const saveBtn = document.getElementById('saveBtn');
  const saveScreenBtn = document.getElementById('saveScreenBtn');
  const bottomBtn = document.getElementById('bottomSaveBtn');
  convertBtn.disabled = true;
  saveBtn.disabled = true;
  saveScreenBtn.disabled = true;
  bottomBtn.disabled = true;
  convertBtn.textContent = mode === 'convert' ? 'Uploading...' : 'Convert to GIF';
  saveBtn.textContent = mode === 'save' ? 'Uploading...' : 'Save Edited Video';
  saveScreenBtn.textContent = mode === 'save' ? 'Uploading...' : 'Save Video';
  bottomBtn.textContent = mode === 'save' ? 'Uploading...' : 'Save Video';
}

function setActionButtonsIdle() {
  const hasVideo = !!selectedFile;
  const convertBtn = document.getElementById('convertBtn');
  const saveBtn = document.getElementById('saveBtn');
  const saveScreenBtn = document.getElementById('saveScreenBtn');
  const bottomBtn = document.getElementById('bottomSaveBtn');
  convertBtn.disabled = !hasVideo;
  saveBtn.disabled = !hasVideo;
  saveScreenBtn.disabled = !hasVideo;
  bottomBtn.disabled = !hasVideo;
  convertBtn.textContent = 'Convert to GIF';
  saveBtn.textContent = 'Save Edited Video';
  saveScreenBtn.textContent = 'Save Video';
  bottomBtn.textContent = 'Save Video';
}

function renderProgressSteps(kind, activeStep, failed) {
  const el = document.getElementById('progressSteps');
  if (!el) return;
  const steps = kind === 'video'
    ? [
        ['uploading', 'Uploading'],
        ['queued', 'Queued'],
        ['setup', 'Setup'],
        ['cutting', 'Cutting'],
        ['joining', 'Joining'],
        ['finalizing', 'Finalizing'],
        ['complete', 'Complete'],
      ]
    : [
        ['uploading', 'Uploading'],
        ['queued', 'Queued'],
        ['setup', 'Setup'],
        ['palette', 'Palette'],
        ['rendering', 'Rendering'],
        ['finalizing', 'Finalizing'],
        ['complete', 'Complete'],
      ];
  const activeIndex = Math.max(0, steps.findIndex(step => step[0] === activeStep));
  el.innerHTML = steps.map((step, index) => {
    const cls = index < activeIndex
      ? 'done'
      : index === activeIndex
        ? (failed ? 'failed' : 'active')
        : '';
    return '<div class="progress-step ' + cls + '">' + step[1] + '</div>';
  }).join('');
}

function stepLabel(kind, stepKey) {
  const labels = kind === 'video'
    ? {
        uploading: 'Uploading',
        queued: 'Queued',
        setup: 'Setup',
        cutting: 'Cutting',
        joining: 'Joining',
        finalizing: 'Finalizing',
        complete: 'Complete',
      }
    : {
        uploading: 'Uploading',
        queued: 'Queued',
        setup: 'Setup',
        palette: 'Palette',
        rendering: 'Rendering',
        finalizing: 'Finalizing',
        complete: 'Complete',
      };
  return labels[stepKey] || 'Processing';
}

function deriveJobStep(job) {
  if (!job) return activeProgressStep || 'uploading';
  if (job.status === 'queued') return 'queued';
  if (job.status === 'done') return 'complete';
  if (job.status === 'failed') return activeProgressStep || 'finalizing';
  const stage = String(job.stage || '').toLowerCase();
  const progress = Number(job.progress || 0);
  if (stage.includes('prob') || stage.includes('start')) return 'setup';
  if ((job.kind || activeJobKind) === 'video') {
    if (progress < 0.15) return 'setup';
    if (progress < 0.62) return 'cutting';
    if (progress < 0.9) return 'joining';
    return 'finalizing';
  }
  if (stage.includes('palette') || progress < 0.45) return 'palette';
  if (progress < 0.9) return 'rendering';
  return 'finalizing';
}

function deriveStagePercent(job, stepKey) {
  if (!job || !stepKey) return null;
  if (job.status === 'done') return 100;
  if (job.status === 'queued') return null;
  const raw = Number(job.progress || 0);
  const ranges = (job.kind || activeJobKind) === 'video'
    ? {
        setup: [0, 0.08],
        cutting: [0.08, 0.62],
        joining: [0.62, 0.9],
        finalizing: [0.9, 1.0],
      }
    : {
        setup: [0, 0.08],
        palette: [0.08, 0.45],
        rendering: [0.45, 0.9],
        finalizing: [0.9, 1.0],
      };
  const range = ranges[stepKey];
  if (!range) return null;
  const start = range[0];
  const end = range[1];
  if (end <= start) return 100;
  const normalized = (raw - start) / (end - start);
  return Math.max(1, Math.min(100, Math.round(normalized * 100)));
}

function showJobResult(job) {
  const panel = document.getElementById('resultPanel');
  const title = document.getElementById('resultTitle');
  const meta = document.getElementById('resultMeta');
  const preview = document.getElementById('resultPreview');
  const viewLink = document.getElementById('resultViewLink');
  const downloadLink = document.getElementById('resultDownloadLink');
  if (!panel || !title || !meta || !preview || !viewLink || !downloadLink || !job) return;

  const viewURL = API + '/jobs/' + job.id + '/view';
  const downloadURL = API + '/jobs/' + job.id + '/download';
  const size = job.result?.output_size ? formatBytes(job.result.output_size) : '';
  const fileName = job.download_name || job.file_name || 'result';
  const kindLabel = job.kind === 'video' ? 'Edited video' : 'GIF';

  title.textContent = kindLabel + ' ready';
  meta.textContent = [fileName, size, job.stage || 'Complete'].filter(Boolean).join(' · ');
  viewLink.href = viewURL;
  viewLink.textContent = job.kind === 'video' ? 'Open Video' : 'Open GIF';
  downloadLink.href = downloadURL;
  downloadLink.textContent = job.kind === 'video' ? 'Download MP4' : 'Download GIF';

  if (job.kind === 'video') {
    preview.innerHTML = '<video controls preload="metadata" src="' + viewURL + '"></video>';
  } else {
    preview.innerHTML = '<img alt="Rendered GIF preview" src="' + viewURL + '">';
  }

  panel.style.display = 'block';
}

function clearResultPanel() {
  const panel = document.getElementById('resultPanel');
  const preview = document.getElementById('resultPreview');
  if (panel) panel.style.display = 'none';
  if (preview) preview.innerHTML = '';
}

function setProgress(pct, label, detail, stepKey, failed, stagePct) {
  if (stepKey) activeProgressStep = stepKey;
  document.getElementById('progressFill').style.width = pct + '%';
  const kind = activeJobKind || 'video';
  let heading = label;
  if (!failed && stepKey) {
    if (stepKey === 'complete') {
      heading = 'Complete (100%)';
    } else if (stagePct != null) {
      heading = stepLabel(kind, stepKey) + ' (' + stagePct + '%)...';
    } else if (stepKey === 'queued') {
      heading = 'Queued...';
    } else {
      heading = stepLabel(kind, stepKey) + '...';
    }
  }
  document.getElementById('progressStatus').textContent = heading;
  document.getElementById('progressPct').textContent = pct ? pct + '%' : '—';
  document.getElementById('progressDetail').textContent = detail;
  renderProgressSteps(activeJobKind || 'video', activeProgressStep || 'uploading', !!failed);
}

function showProgress(show) {
  document.getElementById('progressWrap').style.display = show ? 'block' : 'none';
  if (show) setProgress(0, 'Queued...', 'Waiting for worker...', activeProgressStep || 'uploading');
}

// ── Jobs list ─────────────────────────────────────────────────────────────
async function loadJobs() {
  try {
    const r = await apiFetch(API + '/jobs');
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
    const kind = j.kind === 'video' ? 'edited mp4' : (j.profile?.name || 'gif');
    const pct = j.status === 'done' ? '100%' : (j.status === 'failed' ? 'failed' : (jobProgressPercent(j) + '%'));
    const meta = [kind, j.stage, pct, size].filter(Boolean).join(' · ');
    const buttonLabel = j.kind === 'video' ? '⬇ MP4' : '⬇ GIF';
    return '<div class="job-item">' +
      '<div class="job-status-dot status-' + j.status + '"></div>' +
      '<div class="job-info">' +
        '<div class="job-name">' + escHtml(j.file_name || 'Unknown') + '</div>' +
        '<div class="job-meta">' + escHtml((j.status || '').toUpperCase() + (meta ? ' · ' + meta : '')) + '</div>' +
      '</div>' +
      '<div class="job-actions">' +
        (j.status === 'done'
          ? '<button class="btn-sm btn-download" onclick="triggerDownload(\'' + j.id + '\')">' + buttonLabel + '</button>'
          : '') +
        '<button class="btn-sm btn-delete" onclick="deleteJob(\'' + j.id + '\')">✕</button>' +
      '</div></div>';
  }).join('');
}

async function deleteJob(id) {
  await apiFetch(API + '/jobs/' + id, { method: 'DELETE' });
  if (activeJobID === id) {
    activeJobID = '';
    activeJobKind = '';
    activeProgressStep = '';
    clearInterval(pollInterval);
    setActionButtonsIdle();
    showProgress(false);
    clearResultPanel();
  }
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
    const r = await apiFetch(API + '/config');
    const cfg = await r.json();
    const panel = document.getElementById('configPanel');
    const rows = [
      ['Workers', cfg.queue?.workers],
      ['Max Upload', formatBytes(cfg.server?.max_upload_bytes)],
      ['Upload Dir', cfg.storage?.upload_dir],
      ['Output Dir', cfg.storage?.output_dir],
      ['Share Dir', cfg.storage?.share_dir],
      ['Job Timeout', cfg.queue?.job_timeout_sec + 's'],
      ['File TTL', cfg.storage?.max_age_hours + 'h'],
      ['Auth', cfg.auth?.enabled ? 'enabled' : 'disabled'],
      ['Share TTL', (cfg.sharing?.default_expiry_hours || 0) + 'h'],
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

function formatDateTime(value) {
  if (!value) return 'unknown';
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return 'unknown';
  return d.toLocaleString();
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

function randomID() {
  return Math.random().toString(36).slice(2, 10);
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
