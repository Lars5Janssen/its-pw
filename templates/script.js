document.getElementById('loginButton').addEventListener('click', login);
async function login() {
    const username = document.getElementById('pk-username').value;

    try {
        const response = await fetch('/app/beginLogin', {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username: username })
        });
        if (!response.ok) {
            const msg = await response.json();
            throw new Error('Failed to get login options from server: ' + msg);
        }

        const options = await response.json();

        //JUICE
        const assertionResponse = await SimpleWebAuthnBrowser.startAuthentication(options.publicKey);

        const verificationResponse = await fetch('/app/endLogin', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(assertionResponse)
        });

        const msg = await verificationResponse.json();
        if (verificationResponse.ok) {
            showMessage(msg, false);
        } else {
            showMessage(msg, true)
        }
    } catch (error) {
        showMessage('Error: ' + error.message, true);
    }
}

document.getElementById('registerButton').addEventListener('click', register);
async function register() {
    const username = document.getElementById('pk-username').value;

    try {
        // Get Registration options and Challenge from Server
        const response = await fetch('/app/beginRegistration', {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username: username })
        });

        if (!response.ok) {
            const msg = await response.json();
            throw new Error('User already exists or failed to get reg opt from server: ' + msg);
        }

        const options = await response.json();
        console.log(options)
        const attestationResponse = await SimpleWebAuthnBrowser.startRegistration(options.publicKey);
        console.log(attestationResponse)

        const verificationResponse = await fetch('/app/endRegistration', {
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

document.getElementById('proceed').addEventListener('click', proceed);
async function proceed() {
    const current = window.location.href
    window.location.href = current + "/proceed";
}
