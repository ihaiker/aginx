class RenderLoop {
    constructor(cb, fps = 0) {
        this.currentFps = 0;
        this.isActive = false;
        this.msLastFrame = performance.now();
        this.cb = cb;
        this.totalTime = 0;

        if (fps && typeof fps === 'number' && !Number.isNaN(fps)) {
            this.msFpsLimit = 1000 / fps;
            this.run = () => {
                const currentTime = performance.now();
                const msDt = currentTime - this.msLastFrame;
                this.totalTime += msDt;
                const dt = msDt / 1000;

                if (msDt >= this.msFpsLimit) {
                    this.cb(dt, this.totalTime);
                    this.currentFps = Math.floor(1.0 / dt);
                    this.msLastFrame = currentTime;
                }

                if (this.isActive) window.requestAnimationFrame(this.run);
            };
        } else {
            this.run = () => {
                const currentTime = performance.now();
                const dt = (currentTime - this.msLastFrame) / 1000;
                this.totalTime += (currentTime - this.msLastFrame);
                this.cb(dt, this.totalTime);
                this.currentFps = Math.floor(1.0 / dt);
                this.msLastFrame = currentTime;
                if (this.isActive) window.requestAnimationFrame(this.run);
            };
        }
    }

    changeCb(cb) {
        this.cb = cb;
    }

    start() {
        this.msLastFrame = performance.now();
        this.isActive = true;
        window.requestAnimationFrame(this.run);
        return this;
    }

    stop() {
        this.isActive = false;
        return this;
    }
}

let startTime = performance.now();
const initGl = (canvas, vertexShaderSrc, fragShaderSrc) => {
    const gl = canvas.getContext('webgl2');
    if (!gl) {
        document.write('Please change to a browser which supports WebGl 2.0~');
        return;
    }
    // set background
    gl.clearColor(0, 0, 0, 0.9);

    const vertexShader = gl.createShader(gl.VERTEX_SHADER),
        fragmentShader = gl.createShader(gl.FRAGMENT_SHADER);

    gl.shaderSource(vertexShader, vertexShaderSrc.trim());
    gl.shaderSource(fragmentShader, fragShaderSrc.trim());

    gl.compileShader(vertexShader);
    gl.compileShader(fragmentShader);

    if (!gl.getShaderParameter(vertexShader, gl.COMPILE_STATUS)) {
        console.error(gl.getShaderInfoLog(vertexShader));
        return;
    }

    if (!gl.getShaderParameter(fragmentShader, gl.COMPILE_STATUS)) {
        console.error(gl.getShaderInfoLog(fragmentShader));
        return;
    }

    let program = gl.createProgram();
    gl.attachShader(program, vertexShader);
    gl.attachShader(program, fragmentShader);
    gl.linkProgram(program);

    if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
        console.log(gl.getProgramInfoLog(program));
    }

    gl.useProgram(program);

    return {gl, program};
}
const wrapper = document.querySelector('#wrapper');
const {width: w, height: h} = wrapper.getBoundingClientRect();
const cvs = document.querySelector('#cvs');
cvs.width = w;
cvs.height = h;

const vertexShaderSrc = document.querySelector('#vertex').text.trim();

const fragShaderSrc = document.querySelector('#fragment').text.trim();

const {gl, program} = initGl(cvs, vertexShaderSrc, fragShaderSrc);

gl.enable(gl.DEPTH_TEST);


let vertexBuffer = gl.createBuffer();
let indexBuffer = gl.createBuffer();

gl.bindBuffer(gl.ARRAY_BUFFER, vertexBuffer);
gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1.0, 1.0, -1.0, -1.0, 1.0, -1.0, 1.0, 1.0]), gl.STATIC_DRAW);
gl.vertexAttribPointer(0, 2, gl.FLOAT, false, 0, 0);

gl.enableVertexAttribArray(0);

gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, indexBuffer);
gl.bufferData(gl.ELEMENT_ARRAY_BUFFER, new Uint16Array([1, 0, 2, 3]), gl.STATIC_DRAW);

const uResolution = gl.getUniformLocation(program, 'iResolution');
const {width, height} = cvs.getBoundingClientRect();
gl.uniform2f(uResolution, width, height);
const uTimeIndex = gl.getUniformLocation(program, 'iTime');

new RenderLoop(function (dt, tInMs) {
    gl.uniform1f(uTimeIndex, tInMs / 1000.0);

    gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
    gl.drawElements(gl.TRIANGLE_STRIP, 4, gl.UNSIGNED_SHORT, 0);
}).start();