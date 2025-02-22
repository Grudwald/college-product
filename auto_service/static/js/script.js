function openEditModal(id, date, description, cost) {
    document.getElementById('recordId').value = id;
    document.getElementById('editDate').value = date;
    document.getElementById('editDescription').value = description;
    document.getElementById('editCost').value = cost;

    // Показать модальное окно редактирования
    document.getElementById('editModal').style.display = 'block';
    document.body.classList.add('modal-open');
}

function closeEditModal() {
    document.getElementById('editModal').style.display = 'none';
    document.body.classList.remove('modal-open');
}

function openDeleteConfirmModal(recordId) {
    document.getElementById('deleteRecordId').value = recordId; // Запоминаем ID записи для удаления
    // Показать модальное окно подтверждения удаления
    document.getElementById('deleteConfirmModal').style.display = 'block';
    document.body.classList.add('modal-open');
}

function closeDeleteConfirmModal() {
    document.getElementById('deleteConfirmModal').style.display = 'none';
    document.body.classList.remove('modal-open');
}

// Закрытие модального окна при клике вне него
window.onclick = function(event) {
    var editModal = document.getElementById('editModal');
    var deleteModal = document.getElementById('deleteConfirmModal');
    if (event.target === editModal) {
        closeEditModal();
    } else if (event.target === deleteModal) {
        closeDeleteConfirmModal();
    }
};

