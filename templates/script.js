document.getElementById('impl-loginButton').addEventListener('click', implLogin);

function showImplDetails(sessionId, sessionKey, isError = false) {
    const messageID = document.getElementById('impl-session-id');
    const messageKey = document.getElementById('impl-session-key');
    messageID.textContent = sessionId;
    messageKey.textContent = sessionKey;
    messageID.style.Color = isError ? 'red' : 'green';
    messageKey.style.Color = isError ? 'red' : 'green';
}

function showMessage(message, isError = false) {
    const messageElement = document.getElementById('message');
    messageElement.textContent = message;
    messageElement.style.Color = isError ? 'red' : 'green';
}

async function implLogin() {
    console.log("implLogin");
    var username = document.getElementById('impl-username').value;
    var sharedSecret = document.getElementById('impl-shared-secret').value;

    try {
        var nounceMe = Date.now();
        var crpytNounceMe = encrypt(nounceMe, sharedSecret);
        console.log("NOUNCE");
        console.log(crpytNounceMe);
        const response = await fetch('app/impl/sendLogin', {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username: username, encryptedData: crpytNounceMe})
        });
        if (!response.ok) {
            const msg = await response.json();
            throw new Error('Failed to get login options from server: ' + msg);
        }
        const resposeData = await response.json();
        console.log(resposeData)

    } catch (error) {
        console.log("ERROR");
        showMessage('Error: ' + error.message, true);
    }
}

function encrypt(plaintext, secret) {
    var key = CryptoJS.enc.Utf8.parse(secret);
    let iv = CryptoJS.lib.WordArray.create(key.words.slice(0,4));
    console.log("IV: " + CryptoJS.enc.Base64.stringify(iv));

    // Encrypt
    var cipherText = CryptoJS.AES.encrypt(plaintext, key, {
        iv: iv,
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7
    });
    return cipherText.toString();
}

function decrypt(ciperText, secret, iv) {
    // IV is base64
    let iv1 = CryptoJS.enc.Base64.parse(iv);

    var key = CryptoJS.enc.Utf8.parse(secret);
    var cipherBytes = CryptoJS.enc.Base64.parse(ciperText);

    var decripted = CryptoJS.AES.decrypt({ciphertext: cipherBytes}, key, {
        iv: iv1,
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7
    });
    return decripted.toString(CryptoJS.enc.Utf8);
}

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
