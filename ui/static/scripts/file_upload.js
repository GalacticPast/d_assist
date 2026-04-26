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
  upload_file(files[0]);
  fileInput.value = "";
});

// Handle selected files
fileInput.onchange = (e) => {
  const files = fileInput.files;
  upload_file(files[0]);
  fileInput.value = "";
};

async function upload_file(file) {
  if (!file || file.type !== "application/pdf") {
    alert("Please upload a PDF file.");
    return;
  }

  const formData = new FormData();
  formData.append("syllabus", file);
  // i think this is counter to the tao of datastar. Im too stupid to make this work
  // well actually I think I can make it work now
  // oh well
  const input_element = document.getElementById("is_uploading");
  input_element.value = "show_spinner";
  input_element.dispatchEvent(new Event("input", { bubbles: true }));

  try {
    const upload_url = new URL("/upload", window.location.origin);
    upload_url.searchParams.append("file_path", `${file.name}`);

    let response = await fetch(upload_url);
    const data = await response.json();
    const { file_name: rand_file_path, url: signed_upload_url } = data;

    const { createClient } = supabase;
    const supabase_client = createClient(
      "https://wtpfmvqjwzkwtsvswtmm.supabase.co",
      "sb_publishable_2KZxpcTep54b22QU9jN6Xg_w6v3Erhj",
    );

    const url_obj = new URL(signed_upload_url, window.location.origin);

    const token_from_signed_upload_url = url_obj.searchParams.get("token");

    const { data: uploadData, error: uploadError } =
      await supabase_client.storage
        .from("syllabus_pdf")
        .uploadToSignedUrl(rand_file_path, token_from_signed_upload_url, file);

    if (uploadError) {
      throw new Error(`Upload Failed: ${uploadError.message}`);
    }
    // i think this is counter to the tao of datastar. Im too stupid to make this work
    // well actually I think I can make it work now
    // oh well
    const upload_finished = new URL("/upload_finished", window.location.origin);
    upload_finished.searchParams.append("file_path", `${rand_file_path}`);
    response = await fetch(upload_finished);
  } catch (err) {
    alert(`Error: ${err.message}`);
  } finally {
    input_element.value = "";
    input_element.dispatchEvent(new Event("input", { bubbles: true }));
  }
}
