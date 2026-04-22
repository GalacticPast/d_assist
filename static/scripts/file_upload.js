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
  const files = fileInput.files;
  handleUpload(files[0]);
};
// Add 'async' here
async function handleUpload(file) {
  if (!file || file.type !== "application/pdf") {
    alert("Please upload a PDF file.");
    return;
  }

  const formData = new FormData();
  formData.append("syllabus", file);

  const { createClient } = supabase;
  const cookie = await cookieStore.get("d_assist");

  // 2. Extract the .value
  const token = cookie ? cookie.value : null;
  const supabaseClient = createClient(
    "https://wtpfmvqjwzkwtsvswtmm.supabase.co",
    "sb_publishable_2KZxpcTep54b22QU9jN6Xg_w6v3Erhj",
    {
      accessToken: async () => {
        return token;
      },
    },
  );

  try {
    // 1. Upload to Supabase bucket
    const filePath = `${file.name}`;
    // Now 'await' will work perfectly
    const { data: uploadData, error: uploadError } =
      await supabaseClient.storage.from("syllabus_pdf").upload(filePath, file);

    if (uploadError) {
      throw new Error(`Upload Failed: ${uploadError.message}`);
    }
  } catch (err) {
    console.error(err);
    alert(`Error: ${err.message}`);
  }
}
