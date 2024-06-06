const textarea = document.getElementById('content');

textarea.addEventListener('focus', function() {
    this.style.height = '100px';
    this.setAttribute('placeholder', '');
});

textarea.addEventListener('blur', function() {
    this.style.height = '25px';
    if (this.value.trim() === '') {
        this.setAttribute('placeholder', 'Add a comment');
    }
});
