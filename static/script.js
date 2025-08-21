console.log("Loaded script.js");
const messages = document.getElementById('messages');

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
        body: JSON.stringify({ 'input': prompt, 'model': model })
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
}

async function fetchSessions() {
    console.log("Fetching sessions...");
    const sessionList = document.getElementById('session-list');

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
        li.textContent = `- ${sessionChatHistory[0]['content']}`;
        li.onclick = () => loadSession(sessionID);
        sessionList.appendChild(li);
    }
    
}

async function loadSession(sessionID) {
    currentSessionID = sessionID;

    // Clear the current chat
    messages.innerHTML = '';

    // Fetch the session messages
    const res = await fetch(`/session?key=${sessionID}&format=html`);
    if (!res.ok) {
        console.error(`Error fetching session ${sessionID}:`, res.statusText);
        return;
    }

    const sessionData = await res.json();

    // Process the returned session data
    const sessionChatHistory = sessionData['context']; // This is an array of struct that contain message and role

    for (const message of sessionChatHistory) {
        addMessage(message['role'], message['content']);
    }
}

// Window load event
window.onload = fetchSessions;