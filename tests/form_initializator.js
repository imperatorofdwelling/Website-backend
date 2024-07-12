//Инициализация виджета. Все параметры обязательные.
const payoutsData = new window.PayoutsData({
    type: 'payout',
    account_id: '<Идентификатор_шлюза>', //Идентификатор шлюза (agentId в личном кабинете)
    success_callback: function(data) {
        console.log(data)
    },
    error_callback: function(error) {
        //Обработка ошибок при получении токена карты
    }
});

//Отображение формы в контейнере
payoutsData.render('payout-form')
    .then(() => {
        //Код, который нужно выполнить после отображения формы.
    });