$(function() {
    let divCount = 0;
    $("#search-input").autocomplete({
        source: function (request, response) {
            divCount = 0;
            console.log("Making AJAX request to: /search_autocomplete");
            $.ajax({
                url: "/search_autocomplete?search=" + encodeURIComponent($("#search-input").val()),
                data: {term: request.term},
                success: function (data) {
                    response($.map(data, function (item) {
                        return {
                            label: item.value,
                            value: item.value,
                            type: item.type,
                            profil_picture: item.profil_picture
                        };
                    }));
                }
            });
        },
        classes : {
            "ui-autocomplete": "autocomplete-results"
        },
        minLength: 1,
        select: function (event, ui) {
            if (ui.item.type === "topic") {
                window.location.href = "/posts?topic=" + ui.item.value;
            } else if (ui.item.type === "user") {
                window.location.href = "/user?username=" + ui.item.value;
            }
        }
    }).autocomplete("instance")._renderItem = function(ul, item) {
        divCount++;
        var imgSrc = item.type === "topic" ? "/static/img/topic-logo.png" : item.profil_picture ;
        var colorClass = getColorClass(divCount);
        return $("<li>")
            .append("<div class='rounded-image-div " + colorClass + "'><img class='rounded-image' src='" + imgSrc + "' width='50px' height='50px' /> " + item.label + "</div>")
            .appendTo(ul);
    };
});

    function getColorClass(count) {
        switch (count % 3) {
            case 1:
                return 'color1';
            case 2:
                return 'color2';
            default:
                return 'color3';
        }
    }