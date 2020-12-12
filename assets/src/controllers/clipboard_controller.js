import { Controller } from "stimulus"

export default class extends Controller {
    static targets = [ "source", "copied" ]
    static classes = [ "copied" ]
    copy(e) {
        e.preventDefault()
        this.sourceTarget.select()
        document.execCommand("copy")
        this.copiedTarget.classList.remove(this.copiedClass)
    }
}