import { Controller } from "stimulus"

export default class extends Controller {
    static targets = [ "name","email","magicEmail","password", "confirmPassword", "formError","magic"]

    validateEmail(e) {
        if (!this.emailTarget.validity.valid || this.emailTarget.value === ''){
            this.showFormError("Invalid Email")
            return false;
        }
        return true;
    }

    validateMagicEmail(e) {
        if (!this.magicEmailTarget.validity.valid || this.magicEmailTarget.value === ''){
            this.showFormError("Invalid Email")
            return false;
        }
        return true;
    }


    validateName() {
        if (this.nameTarget.value === ''){
            this.showFormError("Name cannot be empty")
            return false;
        }
        return true;
    }

    validatePassword() {
        if (this.passwordTarget.value === '' || this.passwordTarget.value.length < 8){
            this.showFormError("Minimum 8 character password is required")
            return false;
        }
        return true;
    }

    submitForgetForm(e){
        if (this.validateEmail()){
            return;
        }
        e.preventDefault();
    }

    submitSignupForm(e){
        if (this.validateEmail() && this.validatePassword() && this.validateName()){
            return;
        }
        e.preventDefault();
    }

    submitLoginForm(e){
        if (this.validateEmail() && this.validatePassword()){
            return;
        }
        e.preventDefault();
    }

    submitMagicLoginForm(e){
        if (this.validateMagicEmail()){
            return;
        }
        e.preventDefault();
    }

    submitResetForm(e){
        if (this.validatePassword() &&(this.passwordTarget.value === this.confirmPasswordTarget.value)){
            return;
        }

        this.showFormError("Password's don't match")
        e.preventDefault();
    }

    submitAccountForm(e){
        if (this.validateEmail() && this.validateName()){
            return;
        }
        e.preventDefault();
    }

    showFormError(message){
        this.formErrorTarget.classList.remove("is-hidden")
        this.formErrorTarget.innerHTML = message
    }

    hideFormError(){
        this.formErrorTarget.classList.add("is-hidden")
        e.preventDefault();
    }
}
