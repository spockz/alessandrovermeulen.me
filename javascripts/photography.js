
$(function() {
    $.getJSON("http://api.flickr.com/services/rest/?api_key=6cb7449543a9595800bc0c365223a4e8&extras=url_s,url_m,url_z,url_l&format=json&jsoncallback=flickrshow_jsonp_236309490399&page=1&per_page=50&method=flickr.photos.search&tags=publish-site&user_id=34107923@N05&jsoncallback=?", displayImages);

    function displayImages(data) {

        // Start putting together the HTML string
        var htmlString = "";

        // Now start cycling through our array of Flickr photo details
        $.each(data.photos.photo, function(i,photo){

            // I only want the ickle square thumbnails
            var carousel = $('#myCarousel .carousel-inner');

            // Here's where we piece together the HTML
            var img = $('<img width="100%"/>')
            img.attr('src', photo.url_l.replace('_l.jpg', '_b.jpg'));
            img.attr('title', photo.title);

            var item = $('<div class="item"/>')
            if (i == 0) {
                item.addClass('active');
            }
            var caption = '<div class="carousel-caption"><h4>'+ photo.title + '</h4></div>'
            item.append(img);
            item.append(caption);

            carousel.append(item);
        });
        var a = 54;
        // Pop our HTML in the #images DIV
        $('#myCarousel').carousel();

        // Close down the JSON function call
    }
});