function install() {
    document.title = "安装程序 - MeowBox";
    const container = document.createElement('div');
    container.classList.add('container');
    document.body.appendChild(container);
}
document.addEventListener('DOMContentLoaded', () => {
    install();
});