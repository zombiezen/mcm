// Copyright 2017 The Minimal Configuration Manager Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

extern crate libc;

use std::ffi::CString;
use std::result;
use std::io;
use std::ptr;
use self::libc::{c_char, c_int, c_void, size_t};

#[allow(non_camel_case_types)]
enum lua_State {}

#[allow(non_camel_case_types)]
type lua_KContext = libc::intptr_t;
#[allow(non_camel_case_types)]
type lua_KFunction = extern fn(*mut lua_State, c_int, lua_KContext) -> c_int;
#[allow(non_camel_case_types)]
type lua_Reader = extern fn(*mut lua_State, *mut c_void, *mut size_t) -> *const c_char;

#[link(name = "lua")]
extern {
    fn luaL_newstate() -> *mut lua_State;
    fn lua_close(L: *mut lua_State);
    fn lua_load(L: *mut lua_State, reader: lua_Reader, dt: *mut c_void, chunkname: *const c_char, mode: *const c_char) -> c_int;
    fn luaL_openlibs(L: *mut lua_State);
    fn lua_pcallk(L: *mut lua_State, nargs: c_int, nresults: c_int, msgh: c_int, ctx: lua_KContext, k: Option<lua_KFunction>) -> c_int;
}

#[derive(Clone, Copy, Debug, Eq, Hash, Ord, PartialEq, PartialOrd)]
pub enum Error {
    Runtime,
    Syntax,
    Memory,
    IO,
    Err,
    GcMetamethod,
    Unknown,
}

impl Default for Error {
    fn default() -> Error { Error::Unknown }
}

pub type Result<T> = result::Result<T, Error>;

pub struct State {
    ptr: *mut lua_State,
}

impl State {
    pub fn new() -> State {
        State {ptr: unsafe { luaL_newstate() }}
    }

    pub fn load<R: io::BufRead>(&mut self, chunkname: &str, r: &mut R) -> Result<()> {
        let mut rr = Reader{bufread: r, advance: 0, err: None};
        let retval = {
            let cname = CString::new(chunkname).unwrap();
            unsafe {
                lua_load(self.ptr, readerfunc, &mut rr as *mut _ as *mut c_void, cname.as_ptr(), ptr::null())
            }
        };
        if let Some(_) = rr.err {
            Err(Error::IO)
        } else {
            map_lua_error(retval).map_or(Ok(()), |e| Err(e))
        }
    }

    pub fn load_string(&mut self, s: &str) -> Result<()> {
        let mut arr = s.as_bytes();
        self.load("=(load)", &mut arr)
    }

    pub fn open_libs(&mut self) {
        unsafe {
            luaL_openlibs(self.ptr);
        }
    }

    pub fn pcall(&mut self, nargs: i32, nresults: i32, msgh: i32) -> Result<()> {
        let retval = unsafe { lua_pcallk(self.ptr, nargs, nresults, msgh, 0, None) };
        map_lua_error(retval).map_or(Ok(()), |e| Err(e))
    }
}

impl Drop for State {
    fn drop(&mut self) {
        unsafe {
            lua_close(self.ptr);
        }
    }
}

struct Reader<'a> {
    bufread: &'a mut io::BufRead,
    advance: usize,
    err: Option<io::Error>,
}

extern fn readerfunc(_: *mut lua_State, ptr: *mut c_void, size: *mut size_t) -> *const c_char {
    let r: &mut Reader = unsafe { &mut *(ptr as *mut Reader) };
    if r.advance > 0 {
        r.bufread.consume(r.advance);
    }
    match r.bufread.fill_buf() {
        Ok(buf) => {
            unsafe { *size = buf.len(); }
            r.advance = buf.len();
            buf.as_ptr() as *const c_char
        },
        Err(e) => {
            r.advance = 0;
            r.err = Some(e);
            unsafe { *size = 0; }
            ptr::null()
        },
    }
}

fn map_lua_error(code: c_int) -> Option<Error> {
    match code {
        0 => None,
        2 => Some(Error::Runtime),
        3 => Some(Error::Syntax),
        4 => Some(Error::Memory),
        5 => Some(Error::GcMetamethod),
        6 => Some(Error::Err),
        _ => Some(Error::Unknown),
    }
}
