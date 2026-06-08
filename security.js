// Prevent inspecting HTML source code
document.addEventListener('contextmenu', function(e) {
  e.preventDefault();
});

document.onkeydown = function(e) {
  // Disable F12
  if (e.keyCode === 123) {
    return false;
  }
  // Disable Ctrl+Shift+I, Ctrl+Shift+C, Ctrl+Shift+J, Ctrl+U
  if (e.ctrlKey && e.shiftKey && (e.keyCode === 73 || e.keyCode === 67 || e.keyCode === 74)) {
    return false;
  }
  // Disable Ctrl+U
  if (e.ctrlKey && e.keyCode === 85) {
    return false;
  }
};
