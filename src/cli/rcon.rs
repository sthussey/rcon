use serde_derive::Deserialize;
use toml::de::Error;

#[derive(Deserialize)]
struct Config {
    run: String,
    wd: Option<String>,
}

fn main() {
    let config_parse = parse_config(b"run = '/test'\nwd = '/tmp'\n");

    match config_parse {
        Ok(config) => println!("Running {}", config.run),
        Err(err) => println!("Error parsing config: {}", err),
    };
}

fn parse_config(cfg: &[u8]) -> Result<Config, Error> {
    let config = toml::from_slice(cfg);
    return config;
}
