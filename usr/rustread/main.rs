use std::io;
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
        }
    }
}

fn read_line() -> String {
    let mut line = String::new();
    let stdin = io::stdin(); // We get `Stdin` here.
    let max_count = 10000;
    let mut count = max_count;

    loop {
        let mut buffer = String::new();
        stdin.read_line(&mut buffer).unwrap();
        for c in buffer.chars() {
            
            match c {
                '\x08' => line = rem_last(&line).to_string(),
                '\n' => return line,
                _ => if line.chars().count() < SCREEN_WIDTH - 3 {
                        line.push(c)
                    }
                    
            }
        }
        let mut actual_line = line.clone();
        if count < max_count/2 {
            actual_line.push('_')
        }
        print!("\r> {: <78}", actual_line);
        count = count - 1;
        if count <= 0{
           count = max_count; 
        }
    }
}

fn rem_last(value: &str) -> &str {
    let mut chars = value.chars();
    chars.next_back();
    chars.as_str()
}
