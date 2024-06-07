const mybutton = document.getElementById("back-to-top-btn");

window.onscroll = function() {scrollFunction()};

function scrollFunction() {
    if (document.body.scrollTop > 800 || document.documentElement.scrollTop > 800) {
        mybutton.classList.add("show");
    } else {
        mybutton.classList.remove("show");
    }
}

mybutton.onclick = function() {
    window.scrollTo({ top: 0, behavior: 'smooth' });
}