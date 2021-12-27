use std::io;
use std::io::Read;
use std::io::Write;
use uname_rs::Uname;

const SCREEN_WIDTH: usize = 80;

fn main() -> io::Result<()> {
    let uts = Uname::new()?;	
    println!("This is '{}' version '{}'({})", uts.sysname, uts.release, uts.version);
    loop {
        let line = read_line();
        println!("\r{: <80}", line);
        if line == "secret" {
            println!("Yoo! Secret!")
        } else if line == "exit" {
            return Ok(())
        }
    }
}

fn read_line() -> String {
    let mut line = String::new();
    loop {
        let mut actual_line = line.clone();
        actual_line.push('_');
        print!("\r> {: <77}", actual_line);
        let _a = io::stdout().flush();
        let mut buffer = [0; 1];
        let _n = io::stdin().read(&mut buffer[..]);
        for c in buffer {
            let t = c as char;
            match t {
                '\x08' => line = rem_last(&line).to_string(),
                '\n' => return line,
                _ => if line.chars().count() < SCREEN_WIDTH - 3 {
                        line.push(t)
                    }
            }
        }
    }
}

fn rem_last(value: &str) -> &str {
    let mut chars = value.chars();
    chars.next_back();
    chars.as_str()
}
