console.log("Loaded script.js");
const messages = document.getElementById('messages');

// Crockford's Base32 (no I, L, O, U)
const ENCODING = "0123456789ABCDEFGHJKMNPQRSTVWXYZ";

function encodeTime(ms, len = 10) {
  // ms is a Number; 48 bits fits safely in JS Number (53-bit mantissa)
  let t = ms;
  const out = Array(len);
  for (let i = len - 1; i >= 0; i--) {
    out[i] = ENCODING[t % 32];
    t = Math.floor(t / 32);
  }
  return out.join("");
}

function randomChars(len) {
  // Works in browsers and modern Node (global crypto with getRandomValues)
  const bytes = new Uint8Array(len);
  crypto.getRandomValues(bytes);
  // 256 is divisible by 32, so (byte & 31) is unbiased
  let s = "";
  for (let i = 0; i < len; i++) s += ENCODING[bytes[i] & 31];
  return s;
}

function ulid(now = Date.now()) {
  return encodeTime(now, 10) + randomChars(16);
}

let currentSessionULID = ulid();

// Add a message to the messages area
function addMessage(role, text) {
    const msg = document.createElement('div');
    msg.className = 'msg';

    if (role === 'user') {
        role = 'user';
    } else if (role === 'model') {
        role = 'bot';
    }

    msg.innerHTML = `
    <span class="${role}">${role === 'user' ? 'You' : 'Bot'}:</span>
    <div class="message-body">${DOMPurify.sanitize(text)}</div>
    `;
    messages.appendChild(msg);

    if (window.MathJax) {
        console.log("Rendering this message with MathJax:", text);
        MathJax.typesetPromise([msg]).catch(err => console.error('MathJax error:', err));
    }
    messages.scrollTop = messages.scrollHeight;
}

function addImage(role, encodedImage) {
    const msg = document.createElement('div');
    msg.className = 'msg';
    msg.innerHTML = `<span class="${role}">${role === 'user' ? 'You' : 'Bot'}:</span> <img src="data:image/png;base64,${encodedImage}" alt="Generated Image" style="max-width: 100%; height: auto;">`;
    messages.appendChild(msg);
    messages.scrollTop = messages.scrollHeight;
}

async function send() {
    const input = document.getElementById('input');
    const model = document.getElementById('model').value;
    const prompt = input.value.trim();
    if (!prompt) return;

    addMessage('user', prompt);
    input.value = '';

    const res = await fetch('/chat?format=html', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ 'input': prompt, 'model': model, 'sessionID': currentSessionULID })
    });

    if (!res.ok) {
        addMessage('bot', 'Error: ' + res.statusText);
        return;
    }

    const data = await res.json();
    if (model === 'gemini-2.0-flash-preview-image-generation') {
        addImage('bot', data.response);
    } else {
        addMessage('bot', data.response);
    }

    // Update the session list after sending a message
    fetchSessions();
}

async function fetchSessions() {
    console.log("Fetching sessions...");
    const sessionList = document.getElementById('session-list');
    sessionList.innerHTML = '';

    // Get the all the sessions ID
    const res = await fetch('/session?key=allid');
    if (!res.ok) {
        console.error('Error fetching sessions:', res.statusText);
        return;
    }
    
    const sessions = await res.json();
    // sessions is a hash with the key "keys" and the value being an array of sessionIDs
    keys = sessions['keys'];
    // log the session IDs
    console.log("Fetched session IDs:", keys);
    for (const sessionID of keys) {
        const res = await fetch(`/session?key=${sessionID}`);
        if (!res.ok) {
            console.error(`Error fetching session ${sessionID}:`, res.statusText);
            continue;
        }
        const sessionData = await res.json();

        // Process the returned session data
        const sessionChatHistory = sessionData['context']; // This is an array of struct that contain message and role
        console.log(`Session ${sessionID} chat history:`, sessionChatHistory);

        const li = document.createElement('li');
        li.className = 'session-link';
        li.onclick = () => loadSession(sessionID);

        const linkContent = document.createElement('span');
        linkContent.className = 'session-context';
        linkContent.textContent = `- ${sessionChatHistory[0]['content']}`;

        // Creating a three-dot dropdown container
        const dropdownContainer = document.createElement('div');
        dropdownContainer.className = 'dropdown';

        // Creating the three-dot button
        const dropdownButton = document.createElement('span');
        dropdownButton.className = 'dropdown-button';
        dropdownButton.textContent = '...';
        dropdownButton.onclick = (e) => {
            e.stopPropagation(); // Prevent triggering the link click
            const menu = dropdownContainer.querySelector('.dropdown-menu');
            menu.style.display = menu.style.display === 'block' ? 'none' : 'block';
        }

        // The dropdown menu
        const dropdownMenu = document.createElement('div');
        dropdownMenu.className = 'dropdown-menu';
        dropdownMenu.innerHTML = `
            <div class="dropdown-item" onclick="deleteSession('${sessionID}')">Delete</div>
        `;

        dropdownContainer.appendChild(dropdownButton);
        dropdownContainer.appendChild(dropdownMenu);

        li.appendChild(linkContent);
        li.appendChild(dropdownContainer);
        sessionList.appendChild(li);
    }
}

async function loadSession(sessionID) {
    // Clear the current chat
    messages.innerHTML = '';
    currentSessionULID = sessionID;

    // Fetch the session messages
    const res = await fetch(`/session?key=${sessionID}&format=html`);
    if (!res.ok) {
        console.error(`Error fetching session ${sessionID}:`, res.statusText);
        return;
    }

    const sessionData = await res.json();

    // Change the session id in the cookie
    document.cookie = `Value=${sessionID}; path=/;`;

    // Change model selector to have the model used by that session
    const modelSelector = document.getElementById('model');
    modelSelector.value = sessionData['model'];

    // Process the returned session data
    const sessionChatHistory = sessionData['context']; // This is an array of struct that contain message and role

    for (const message of sessionChatHistory) {
        addMessage(message['role'], message['content']);
    }
}

async function startNewSession() {
    // Clear the current chat
    messages.innerHTML = '';
    currentSessionULID = null;

    // Generate a new session ID
    currentSessionULID = ulid();

    // Reset the model selector to the default value
    const modelSelector = document.getElementById('model');
    modelSelector.value = 'gemma-3-27b-it';

}

async function deleteSession(sessionID) {
    if (!sessionID) {
        console.error("No session ID provided for deletion.");
        return;
    }

    const res = await fetch(`/session?key=${sessionID}`, {
        method: 'DELETE'
    });

    if (!res.ok) {
        console.error(`Error deleting session ${sessionID}:`, res.statusText);
        return;
    }

    console.log(`Session ${sessionID} deleted successfully.`);
    // If the deleted session is the current one, start a new session
    if (currentSessionULID === sessionID) {
        startNewSession();
    }

    // Refresh the session list
    fetchSessions();
}

// Window load event
window.onload = fetchSessions;