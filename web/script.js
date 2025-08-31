document.getElementById('searchBtn').addEventListener('click', async () => {
    const orderId = document.getElementById('orderId').value.trim();
    const resultElem = document.getElementById('result');

    if (!orderId) {
        resultElem.textContent = 'Введите order_uid!';
        return;
    }

    try {
        const res = await fetch(`/orders/${orderId}`);
        if (!res.ok) {
            resultElem.textContent = `Ошибка: ${res.statusText}`;
            return;
        }
        const data = await res.json();
        resultElem.textContent = JSON.stringify(data, null, 2);
    } catch (err) {
        resultElem.textContent = `Ошибка запроса: ${err.message}`;
    }
});
