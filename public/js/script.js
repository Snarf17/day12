function submitData(){
    let name = document.getElementById("name").value
    let email = document.getElementById("email").value
    let number = document.getElementById("number").value
    let subject = document.getElementById("subject").value
    let massage = document.getElementById("massage").value
    
    console.log(name, email, number, subject, massage);
   
    // let form = document.forms["name"]["email"]["number"]["subject"];
    if (name == "") {
        return false;
    }else if (email == "") {
        return false;
    }else if (number == "") {
        return false;
    }else if(subject == ""){
       return false;
    }

    let link = document.createElement("a");
    link.href = `mailto:${name}?subject=${subject}&body=Hallo nama saya ${name}`;
    link.click();
}

