const dropzone = document.getElementById("dropzone");
const fileInput = document.getElementById("file-input");
const statusText = document.getElementById("status-text");
const buttons = document.querySelectorAll(".select-btn");

// Trigger file dialog
buttons.forEach((btn) => {
  btn.addEventListener("click", () => {
    fileInput.click();
  });
});

// Highlighting the dropzone
["dragenter", "dragover"].forEach((eventName) => {
  dropzone.addEventListener(
    eventName,
    (e) => {
      e.preventDefault();
      dropzone.classList.add("active");
    },
    false,
  );
});

["dragleave", "drop"].forEach((eventName) => {
  dropzone.addEventListener(
    eventName,
    (e) => {
      e.preventDefault();
      dropzone.classList.remove("active");
    },
    false,
  );
});

// Handle dropped files
dropzone.addEventListener("drop", (e) => {
  const files = e.dataTransfer.files;
  handleUpload(files[0]);
});

// Handle selected files
fileInput.onchange = (e) => {
  handleUpload(e.target.files[0]);
};

function handleUpload(file) {
  if (!file || file.type !== "application/pdf") {
    alert("Please upload a PDF file.");
    return;
  }

  statusText.innerText = `Uploading: ${file.name}...`;

  const formData = new FormData();
  formData.append("syllabus", file);
}
