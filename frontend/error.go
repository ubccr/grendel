package frontend

func ToastHtml(m, t string) string {
	if t == "success" {
		return "<div class='rounded-lg border border-neutral-400 bg-white p-4 shadow-lg z-100'><h1 class='text-green-600'>" + m + "</h1></div>"
	} else {
		return "<div class='rounded-lg border border-neutral-400 bg-white p-4 shadow-lg z-100'><h1 class='text-red-600'>" + m + "</h1></div>"
	}
}
