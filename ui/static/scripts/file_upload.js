const dropzone = document.getElementById("dropzone");
const statusText = document.getElementById("status-text");

const file_inputs = document.querySelectorAll(".file-input");
const badges = document.querySelectorAll(".badge");
const syllabus_pdf_file_input = document.getElementById("Syllabus-upload");
const syllabus_upload_btn = document.getElementById("Syllabus-upload-btn");

const slides_pdf_file_input = document.getElementById("Slides-upload");
const slides_upload_btn = document.getElementById("Slides-upload-btn");

if (syllabus_upload_btn && syllabus_pdf_file_input) {
  syllabus_upload_btn.addEventListener("click", () => {
    syllabus_pdf_file_input.click();
  });
  syllabus_pdf_file_input.addEventListener("change", (e) => {
    if (syllabus_pdf_file_input.files.length > 0) {
      upload_file("syllabus-upload-btn", syllabus_pdf_file_input.files[0]);
      syllabus_pdf_file_input.value = "";
    }
  });
} else {
  console.log("wtf");
}

if (slides_upload_btn && slides_pdf_file_input) {
  slides_upload_btn.addEventListener("click", () => {
    slides_pdf_file_input.click();
  });
  slides_pdf_file_input.addEventListener("change", (e) => {
    if (slides_pdf_file_input.files.length > 0) {
      upload_file("slides-upload-btn", slides_pdf_file_input.files[0]);
      slides_pdf_file_input.value = "";
    }
  });
} else {
  console.log("wtf");
}

if (dropzone) {
  ["dragenter", "dragover"].forEach((eventName) => {
    dropzone.addEventListener(eventName, (e) => {
      e.preventDefault();
      dropzone.classList.add("active");
    });
  });

  ["dragleave", "drop"].forEach((eventName) => {
    dropzone.addEventListener(eventName, (e) => {
      e.preventDefault();
      dropzone.classList.remove("active");
    });
  });

  dropzone.addEventListener("drop", (e) => {
    const files = e.dataTransfer.files;
    if (files.length > 0) {
      //@hack: this is a hack I need to make it more robust
      if (syllabus_pdf_file_input) {
        upload_file("syllabus-upload-btn", files[0]);
      }
    }
  });
}

async function upload_file(button_id, file) {
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
    upload_url.searchParams.append("button_id", button_id);

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

    const bucketName =
      button_id === "syllabus-upload-btn" ? "syllabus_pdf" : "slides_pdf";

    const { data: upload_data, error: upload_error } =
      await supabase_client.storage
        .from(bucketName)
        .uploadToSignedUrl(rand_file_path, token_from_signed_upload_url, file);

    if (upload_error) {
      throw new Error(`Upload Failed: ${upload_error.message}`);
    }
    // i think this is counter to the tao of datastar. Im too stupid to make this work
    // well actually I think I can make it work now
    // oh well
    let uuid = badges[0].id;
    const upload_finished = new URL("/upload_finished", window.location.origin);
    upload_finished.searchParams.append("button_id", button_id);
    upload_finished.searchParams.append("file_path", `${rand_file_path}`);
    upload_finished.searchParams.append("id", `${uuid}`);
    response = await fetch(upload_finished);
  } catch (err) {
    alert(`Error: ${err.message}`);
  } finally {
    input_element.value = "";
    input_element.dispatchEvent(new Event("input", { bubbles: true }));
  }
}
