const evtSource = new EventSource("/events");
const apologyDiv = document.getElementById("apologies");

evtSource.onmessage = function (event) {
  try {
    const data = JSON.parse(event.data);
    if (data.type === "apology") {
      const bubble = document.createElement("div");
      bubble.className = "p-3 my-2 bg-pink-100 border border-pink-300 rounded-xl animate-fade-in";
      bubble.textContent = data.text;
      apologyDiv.appendChild(bubble);
    }
  } catch (err) {
    console.error("Error parsing SSE:", err);
  }
};

const style = document.createElement("style");
style.textContent = `
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}
.animate-fade-in {
  animation: fadeIn 0.5s ease-out;
}
`;
document.head.appendChild(style);

// --- Handle form submission via Fetch ---
const form = document.getElementById("complaintForm");
const statusText = document.getElementById("status");

form.addEventListener("submit", async (e) => {
  e.preventDefault(); // stop page reload
  const formData = new FormData(form);
  const complaint = formData.get("complaint");

  statusText.textContent = "🕐 Sending your complaint...";
  
  try {
    const response = await fetch("/complain", {
      method: "POST",
      body: formData,
    });
    if (response.ok) {
      statusText.textContent = "✅ Complaint sent! Awaiting AI apology...";
      form.reset();
    } else {
      statusText.textContent = "❌ Failed to send complaint.";
    }
  } catch (err) {
    console.error(err);
    statusText.textContent = "❌ Network error.";
  }
});
