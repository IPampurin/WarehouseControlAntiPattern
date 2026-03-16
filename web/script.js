// Общая функция уведомлений
function showNotification(message, isError = false) {
    const notif = document.getElementById('notification');
    if (!notif) return;
    notif.textContent = message;
    notif.className = 'notification' + (isError ? ' error' : '');
    notif.style.display = 'block';
    setTimeout(() => {
        notif.style.display = 'none';
    }, 3000);
}

// Получение текущей роли из localStorage
function getRole() {
    return localStorage.getItem('userRole');
}

// Загрузка списка товаров (используется на всех страницах)
async function loadItems() {
    try {
        const response = await fetch('/items');
        if (!response.ok) throw new Error('Ошибка загрузки');
        const items = await response.json();
        renderItems(items);
    } catch (err) {
        showNotification('Ошибка загрузки: ' + err.message, true);
    }
}

// Отрисовка товаров с учётом роли
function renderItems(items) {
    const container = document.getElementById('itemsContainer');
    if (!container) return;

    const role = getRole();

    // Сохраняем заголовок
    const header = container.querySelector('.item-header');
    container.innerHTML = '';
    if (header) {
        container.appendChild(header);
    } else {
        const newHeader = document.createElement('div');
        newHeader.className = 'item-header';
        newHeader.innerHTML = `
            <span class="item-id">ID</span>
            <span class="item-name">Название</span>
            <span class="item-quantity">Количество</span>
            <span class="item-price">Цена</span>
            ${role !== 'viewer' ? '<span class="item-actions-header">Действия</span>' : ''}
        `;
        container.appendChild(newHeader);
    }

    items.forEach(item => {
        const row = document.createElement('div');
        row.className = 'item-row';
        row.dataset.id = item.id;

        let actionsHtml = '';
        if (role !== 'viewer') {
            actionsHtml += `<button class="edit-btn" data-id="${item.id}" data-name="${item.name}" data-quantity="${item.quantity}" data-price="${item.price}">Редактировать</button>`;
            if (role === 'admin') {
                actionsHtml += `<button class="delete-btn" data-id="${item.id}">Удалить</button>`;
                actionsHtml += `<button class="history-btn" data-id="${item.id}">История</button>`;
            }
        }

        row.innerHTML = `
            <span class="item-id">${item.id}</span>
            <span class="item-name">${item.name}</span>
            <span class="item-quantity">${item.quantity}</span>
            <span class="item-price">${item.price}</span>
            ${role !== 'viewer' ? `<div class="item-actions">${actionsHtml}</div>` : ''}
        `;

        container.appendChild(row);
    });

    // Привязываем обработчики
    document.querySelectorAll('.edit-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const id = e.target.dataset.id;
            const name = e.target.dataset.name;
            const quantity = e.target.dataset.quantity;
            const price = e.target.dataset.price;
            if (typeof window.editItem === 'function') {
                window.editItem(id, name, quantity, price);
            }
        });
    });

    document.querySelectorAll('.delete-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const id = e.target.dataset.id;
            if (typeof window.deleteItem === 'function') {
                window.deleteItem(id);
            }
        });
    });

    document.querySelectorAll('.history-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const id = e.target.dataset.id;
            if (typeof window.showItemHistory === 'function') {
                window.showItemHistory(id);
            }
        });
    });
}