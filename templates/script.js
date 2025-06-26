document.getElementById('registerButton').addEventListener('click', register);
document.getElementById('loginButton').addEventListener('click', login);

function showMessage(message, isError = false) {
    const messageElement = document.getElementById('message');
    messageElement.textContent = message;
    messageElement.style.Color = isError ? 'red' : 'green';
}

async function register() {
    const username = document.getElementById('pk-username').value;

    try {
        // Get Registration options and Challenge from Server
        const response = await fetch('TODO DEFINE ENDPOINT REGISTER START', {
            method: 'POST', headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({username: username})
        });

        if (!response.ok) {
            const msg = await response.json();
            throw new Error('User already exists or failed to get reg opt from server: ' + msg);
        }

        const options = await response.json();
        const attestationResponse = await SimpleWebAuthnBrowser.startRegistrtion(options.publicKey);

        const verificationResponse = await fetch('TODO DEFINE ENDPOINT REGISTER FINISH', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(attestationResponse)
        });


        const msg = await verificationResponse.json();
        if (verificationResponse.ok) {
            showMessage(msg, false);
        } else {
            showMessage(msg, true);
        }
    } catch
        (error) {
        showMessage('Error: ' + error.message, true);
    }
}
