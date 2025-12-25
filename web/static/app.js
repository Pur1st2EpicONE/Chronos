const apiBase = '/api/v1/notify';

document.addEventListener('DOMContentLoaded', () => {
  const channelSelect = document.getElementById('channel');
  const emailFields = document.getElementById('emailFields');
  const createForm = document.getElementById('createForm');
  const tbody = document.getElementById('notifications');
  const sendAtInput = document.getElementById('sendAt');

  const now = new Date();
  const pad = (num) => String(num).padStart(2, '0');
  const offset = -now.getTimezoneOffset();
  const sign = offset >= 0 ? '+' : '-';
  const hoursOffset = pad(Math.floor(Math.abs(offset) / 60));
  const minutesOffset = pad(Math.abs(offset) % 60);
  const localIso = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}T${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}${sign}${hoursOffset}:${minutesOffset}`;
  sendAtInput.value = localIso;

  channelSelect.addEventListener('change', () => {
    emailFields.style.display = channelSelect.value.toLowerCase() === 'email' ? 'block' : 'none';
  });
  channelSelect.dispatchEvent(new Event('change'));

  async function loadNotifications() {
    try {
      const res = await fetch(apiBase);
      if (!res.ok) return;
      const data = await res.json();

      let list = [];
      if (Array.isArray(data)) list = data;
      else if (Array.isArray(data.notifications)) list = data.notifications;
      else if (Array.isArray(data.results)) list = data.results;
      else if (Array.isArray(data.data)) list = data.data;
      else return;

      list.forEach(n => {
        const id = getId(n);
        const status = getStatus(n);
        if (id != null) addOrUpdateNotificationRow(id, status);
      });

      const title = document.getElementById('notificationsTitle');
      const wrapper = document.getElementById('notificationsWrapper');
      if (tbody.children.length > 0) {
        title.style.display = 'block';
        wrapper.style.display = 'block';
      } else {
        title.style.display = 'none';
        wrapper.style.display = 'none';
      }
    } catch (err) {
      console.error('Failed to load notifications:', err);
    }
  }

  function getId(n) {
    if (n == null) return null;
    return n.id ?? n.ID ?? n.Id ?? n.ID_ ?? null;
  }

  function getStatus(n) {
    if (n == null) return 'unknown';
    return n.status ?? n.Status ?? n.state ?? 'unknown';
  }

  function addOrUpdateNotificationRow(id, status) {
    const title = document.getElementById('notificationsTitle');
    const wrapper = document.getElementById('notificationsWrapper');

    if (title.style.display === 'none') title.style.display = 'block';
    if (wrapper.style.display === 'none') wrapper.style.display = 'block';

    const existingRow = document.getElementById(`notif-${id}`);
    if (existingRow) {
      if (existingRow.cells && existingRow.cells[1]) {
        existingRow.cells[1].innerText = status;
        const btn = existingRow.querySelector('button');
        if (btn) btn.disabled = status === 'canceled';
      }
      return;
    }

    const tr = document.createElement('tr');
    tr.id = `notif-${id}`;

    const tdId = document.createElement('td');
    tdId.innerText = id;

    const tdStatus = document.createElement('td');
    tdStatus.innerText = status;

    const tdAction = document.createElement('td');
    const btn = document.createElement('button');
    btn.type = 'button';
    btn.innerText = 'Cancel';
    btn.disabled = status === 'canceled';
    btn.addEventListener('click', () => cancelNotification(id, btn));
    tdAction.appendChild(btn);

    tr.appendChild(tdId);
    tr.appendChild(tdStatus);
    tr.appendChild(tdAction);

    tbody.appendChild(tr);
  }

  createForm.addEventListener('submit', async (e) => {
    e.preventDefault();

    const channel = channelSelect.value.toLowerCase();
    const payload = {
      channel,
      message: document.getElementById('message').value,
      send_at: sendAtInput.value
    };

    if (channel === 'email') {
      payload.subject = document.getElementById('subject').value;
      payload.send_to = document.getElementById('sendTo').value
        .split(',')
        .map(s => s.trim())
        .filter(Boolean);
      if (!payload.send_to.length) { alert('Email address is required'); return; }
      if (!payload.subject) { alert('Subject is required'); return; }
    }

    try {
      const res = await fetch(apiBase, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      const data = await res.json();

      const createdId = data.result ?? data.id ?? data.ID ?? (data.created && (data.created.id ?? data.created.ID)) ?? null;
      const createdStatus = data.status ?? 'created';

      if (createdId) {
        addOrUpdateNotificationRow(createdId, createdStatus);
        alert(`Created with ID: ${createdId}`);
        channelSelect.dispatchEvent(new Event('change'));
      } else {
        let errorMsg = 'Unknown error';
        if (data.error) errorMsg = data.error;
        else if (typeof data === 'string') errorMsg = data;
        alert(errorMsg);
      }
    } catch (err) {
      alert('Network error: ' + err);
    }
  });

  async function cancelNotification(id, button) {
    try {
      const res = await fetch(`${apiBase}?id=${encodeURIComponent(id)}`, { method: 'DELETE' });
      const data = await res.json();
      const ok = data.result ?? data.ok ?? data.success ?? false;
      if (ok) {
        button.disabled = true;
        const tr = document.getElementById(`notif-${id}`);
        if (tr && tr.cells && tr.cells[1]) tr.cells[1].innerText = 'canceled';
      } else {
        let errorMsg = 'Unknown error';
        if (data.error) errorMsg = data.error;
        else if (data.message) errorMsg = data.message;
        else if (typeof data === 'string') errorMsg = data;
        alert(errorMsg);
      }
    } catch (err) {
      alert('Network error: ' + err);
    }
  }

  window.addOrUpdateNotificationRow = addOrUpdateNotificationRow;
  window.cancelNotification = cancelNotification;

  loadNotifications();
});
