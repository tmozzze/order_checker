document.addEventListener('DOMContentLoaded', function() {
  console.log("Order Checker frontend loaded successfully!");
  
  const searchBtn = document.getElementById('searchBtn');
  const orderIdInput = document.getElementById('orderId');
  const resultDiv = document.getElementById('result');

  searchBtn.addEventListener('click', handleSearch);
  
  orderIdInput.addEventListener('keypress', function(e) {
    if (e.key === 'Enter') {
      handleSearch();
    }
  });

  async function handleSearch() {
    const orderId = orderIdInput.value.trim();
    
    if (!orderId) {
      showError('Пожалуйста, введите order_uid.');
      return;
    }

    showLoading();

    try {
      const response = await fetch(`/orders/${orderId}`);
      
      if (!response.ok) {
        if (response.status === 404) {
          throw new Error('Заказ не найден');
        } else {
          throw new Error(`Ошибка сервера: ${response.status}`);
        }
      }

      const orderData = await response.json();
      displayOrder(orderData);
    } catch (error) {
      showError(error.message);
    }
  }

  function displayOrder(order) {
    // Форматируем дату для красивого отображения
    const formattedDate = order.date_created ? 
      new Date(order.date_created).toLocaleString('ru-RU') : 
      'Не указана';

    resultDiv.innerHTML = `
      <div class="card">
        <div class="section">
          <h3>Информация о заказе</h3>
          <p><strong>Order UID:</strong> ${order.order_uid || 'Не указан'}</p>
          <p><strong>Track Number:</strong> ${order.track_number || 'Не указан'}</p>
          <p><strong>Entry:</strong> ${order.entry || 'Не указано'}</p>
          <p><strong>Дата создания:</strong> ${formattedDate}</p>
          <p><strong>Locale:</strong> ${order.locale || 'Не указан'}</p>
          <p><strong>Customer ID:</strong> ${order.customer_id || 'Не указан'}</p>
        </div>

        ${order.delivery ? `
        <div class="section">
          <h3>Доставка</h3>
          <p><strong>Имя:</strong> ${order.delivery.name || 'Не указано'}</p>
          <p><strong>Телефон:</strong> ${order.delivery.phone || 'Не указан'}</p>
          <p><strong>Город:</strong> ${order.delivery.city || 'Не указан'}</p>
          <p><strong>Адрес:</strong> ${order.delivery.address || 'Не указан'}</p>
          <p><strong>Регион:</strong> ${order.delivery.region || 'Не указан'}</p>
          <p><strong>Email:</strong> ${order.delivery.email || 'Не указан'}</p>
        </div>
        ` : ''}

        ${order.payment ? `
        <div class="section">
          <h3>Оплата</h3>
          <p><strong>Транзакция:</strong> ${order.payment.transaction || 'Не указана'}</p>
          <p><strong>Валюта:</strong> ${order.payment.currency || 'Не указана'}</p>
          <p><strong>Провайдер:</strong> ${order.payment.provider || 'Не указан'}</p>
          <p><strong>Сумма:</strong> ${order.payment.amount ? order.payment.amount + ' ₽' : 'Не указана'}</p>
          <p><strong>Доставка:</strong> ${order.payment.delivery_cost ? order.payment.delivery_cost + ' ₽' : 'Не указана'}</p>
        </div>
        ` : ''}

        ${order.items && order.items.length > 0 ? `
        <div class="section">
          <h3>Товары (${order.items.length})</h3>
          <table>
            <thead>
              <tr>
                <th>Название</th>
                <th>Цена</th>
                <th>Скидка</th>
                <th>Размер</th>
                <th>Общая цена</th>
                <th>Бренд</th>
              </tr>
            </thead>
            <tbody>
              ${order.items.map(item => `
                <tr>
                  <td>${item.name || 'Не указано'}</td>
                  <td>${item.price ? item.price + ' ₽' : '0 ₽'}</td>
                  <td>${item.sale || '0'}</td>
                  <td>${item.size || '0'}</td>
                  
                  <td>${item.total_price ? item.total_price + ' ₽' : '0 ₽'}</td>
                  <td>${item.brand || 'Не указан'}</td>
                </tr>
              `).join('')}
            </tbody>
          </table>
        </div>
        ` : '<div class="section"><p>Нет товаров в заказе</p></div>'}
      </div>
    `;
  }

  function showLoading() {
    resultDiv.innerHTML = '<div class="loading">Поиск заказа...</div>';
  }

  function showError(message) {
    resultDiv.innerHTML = `<div class="error">${message}</div>`;
  }
});