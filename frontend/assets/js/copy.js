
export function copyToClipboard(content) {
    navigator.clipboard.writeText(content).then(() => {
        Swal.fire({
            icon: "success",
            title: "Copied!",
            showConfirmButton: false,
            timer: 1500,
            toast: true,
            position: "top",
            background: "#222",
            color: "#FFD700",
            customClass: {
                popup: "small-swal"
            }
        });
    }).catch(err => {
        Swal.fire({
            icon: "error",
            title: "Oops!",
            text: "Failed to copy text.",
            showConfirmButton: false,
            timer: 1500
        });
        console.error("Failed to copy text: ", err);
    });
}