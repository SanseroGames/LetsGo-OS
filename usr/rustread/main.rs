use std::io;
use std::process;


fn main() -> io::Result<()> {
    let mut buffer = String::new();
    let stdin = io::stdin(); // We get `Stdin` here.
    loop {
        stdin.read_line(&mut buffer)?;
        let mut total: i128 = 0;
        for i in 1..1000 {
            for j in 1..1000 {
                total += (i * j) + (i * j);
            };
        };
        println!("My pid is {}", process::id());
        println!("The total is {:?}", total);
        println!("{}", buffer);
    }
}
