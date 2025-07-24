// Settings page JavaScript functions

// Generate new API token
function generateNewToken() {
    const tokenName = prompt("Podaj nazwę dla nowego tokenu API:");
    if (!tokenName) {
        return;
    }

    fetch('/api-tokens', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            name: tokenName
        })
    })
    .then(response => response.json())
    .then(data => {
        if (data.token) {
            // Show the token to the user
            const token = data.token;
            const message = `Token został wygenerowany!\n\nNazwa: ${tokenName}\nToken: ${token}\n\nSkopiuj ten token i zapisz go w bezpiecznym miejscu. Nie będzie możliwe jego ponowne wyświetlenie.`;
            alert(message);
            
            // Trigger success event
            document.body.dispatchEvent(new Event('token-generated'));
            
            // Reload the page to show the new token
            window.location.reload();
        } else {
            alert('Błąd podczas generowania tokenu: ' + (data.message || 'Nieznany błąd'));
        }
    })
    .catch(error => {
        console.error('Error:', error);
        alert('Błąd podczas generowania tokenu');
    });
}

// Delete API token
function deleteToken(tokenId) {
    if (!confirm('Czy na pewno chcesz usunąć ten token? Ta operacja nie może być cofnięta.')) {
        return;
    }

    fetch(`/api/tokens/${tokenId}`, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json',
        }
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            // Trigger success event
            document.body.dispatchEvent(new Event('token-deleted'));
            
            // Reload the page to update the token list
            window.location.reload();
        } else {
            alert('Błąd podczas usuwania tokenu: ' + (data.message || 'Nieznany błąd'));
        }
    })
    .catch(error => {
        console.error('Error:', error);
        alert('Błąd podczas usuwania tokenu');
    });
} 