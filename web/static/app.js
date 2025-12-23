const apiBase = '/api/v1/notify';
const channelSelect = document.getElementById('channel');
const emailFields = document.getElementById('emailFields');

channelSelect.addEventListener('change', () => {
  emailFields.style.display = channelSelect.value.toLowerCase() === 'email' ? 'block' : 'none';
});
channelSelect.dispatchEvent(new Event('change'));

document.getElementById('createForm').addEventListener('submit', async (e) => {
  e.preventDefault();

  const channel = channelSelect.value.toLowerCase();
  const payload = {
    channel,
    message: document.getElementById('message').value,
    send_at: document.getElementById('sendAt').value
  };

  if (channel === 'email') {
    payload.subject = document.getElementById('subject').value;
    payload.send_to = document.getElementById('sendTo').value.split(',').map(s=>s.trim()).filter(s=>s);
    if (!payload.send_to.length) { alert("email address is required"); return; }
    if (!payload.subject) { alert("subject is required"); return; }
  }

  const res = await fetch(apiBase, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  const data = await res.json();
  if (data.result) {
    addNotificationRow(data.result, "created");
    alert(`Created with ID: ${data.result}`);
  } else {
    alert(JSON.stringify(data.error));
  }
});

function addNotificationRow(id, status) {
  const tbody = document.getElementById('notifications');
  const tr = document.createElement('tr');
  tr.id = `notif-${id}`;
  tr.innerHTML = `
    <td>${id}</td>
    <td>${status}</td>
    <td><button onclick="cancelNotification('${id}', this)">Cancel</button></td>
  `;
  tbody.appendChild(tr);
}

async function cancelNotification(id, button) {
  const res = await fetch(`${apiBase}?id=${id}`, { method: 'DELETE' });
  const data = await res.json();
  if (data.result) {
    button.disabled = true;
    const tr = document.getElementById(`notif-${id}`);
    if (tr) tr.cells[1].innerText = "canceled";
  } else {
    alert(JSON.stringify(data.error));
  }
}
