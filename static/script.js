const messages = document.getElementById('messages');

// Add a message to the messages area
function addMessage(role, text) {
    const msg = document.createElement('div');
    msg.className = 'msg';

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