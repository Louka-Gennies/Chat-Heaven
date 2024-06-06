document.addEventListener("DOMContentLoaded", function() {
    var dropZone = document.getElementById('drop-zone');
    var fileInput = document.getElementById('profile_picture');
    var preview = document.getElementById('preview');
    var text = document.getElementById('text');

    preview.style.display = 'none';
    text.style.display = 'block';

    dropZone.ondrop = function(e) {
        e.preventDefault();
        this.className = 'upload-drop-zone';

        fileInput.files = e.dataTransfer.files;

        var reader = new FileReader();

        reader.onload = function() {
            preview.src = reader.result;
            preview.style.display = 'block';
            text.style.display = 'none';
        }

        if (fileInput.files[0]) {
            reader.readAsDataURL(fileInput.files[0]);
        }
    }

    dropZone.ondragover = function() {
        this.className = 'upload-drop-zone drop';
        return false;
    }

    dropZone.ondragleave = function() {
        this.className = 'upload-drop-zone';
        return false;
    }
});