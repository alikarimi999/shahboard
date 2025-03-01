export function showErrorMessage(txtMsg) {
    if ($('#error').length) return; // Avoid duplicate messages

    const message = $('<div id="error" class="error-message"></div>').text(txtMsg);
    $('body').append(message);

    // Animate the message
    message.fadeIn(300).delay(2000).fadeOut(500, function () {
        $(this).remove();
    });
}
